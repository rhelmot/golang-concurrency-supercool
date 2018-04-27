package silk

import (
    "os"
    "time"
    "bytes"
    "net/http"
    "encoding/gob"
)

type taskWithId struct {
    TaskId int
    Task Task
}

// main server entrypoint
func (self *Server) Serve() (chan ClientCaps, chan Task) {
    if self.serving {
        panic("Called Serve() on already serving server!")
    }
    self.serving = true

    self.nextTaskId = 1
    self.nextNodeId = 1

    self.taskQueue = make(chan taskWithId)
    self.rememberedTasks = make(chan Task)
    self.nodeEvents = make(chan ClientCaps)
    self.taskProgressMap = make(map[int]chan Task)
    self.nodeHeartbeatMap = make(map[int]chan int)

    if self.NodeTimeout == 0 {
        self.NodeTimeout = time.Duration(60 * time.Second)
    }

    if self.ServerId == 0 {
        self.ServerId = int(time.Now().UnixNano())
    }

    mux := http.NewServeMux()
    mux.Handle("/sync", self)
    mux.Handle("/download", http.FileServer(downloadClient{}))

    srv := &http.Server{
        Addr: self.Listen,
        Handler: mux,
    }
    go func() {panic(srv.ListenAndServe())}()

    return self.nodeEvents, self.rememberedTasks
}

// allows for downloading the current binary via GET /download
type downloadClient struct{}
func (self downloadClient) Open(name string) (http.File, error) {
    return os.Open(os.Args[0])
}

// main entry point - responds to POST /sync
func (self *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var oldTask, newTask taskWithId
    var syncReq SyncRequest
    var syncResp SyncResponse

    var heartbeat chan int
    var nodeId int
    var ok bool
    var err error

    var taskOnWire, remembering, sendNewTask bool

    // Step 1: Validate method
    if r.Method != "POST" {
        http.Error(w, "Bad method - must only POST to /sync", 400)
        return
    }

    // Step 2: Receive SyncRequest
    d := gob.NewDecoder(r.Body)
    err = d.Decode(&syncReq)
    if err != nil {
        http.Error(w, "Could not decode SyncRequest", 400)
        return
    }

    // Step 3: Enforce api versioning
    if syncReq.Version != self.Version {
        buf := bytes.Buffer{}
        e := gob.NewEncoder(&buf)
        syncResp = SyncResponse{self.Version, -1, -1, "Must upgrade"}
        err = e.Encode(&syncResp)
        if err != nil {
            http.Error(w, "Could not encode SyncResponse for upgrade..?", 500)
        } else {
            _, err = buf.WriteTo(w)
            if err != nil {
                // TODO: log nasty message
            }
        }
        return
    }

    // Step 4: Node pool membership
    taskOnWire = false // are we receiving a task?
    remembering = false // are we receiving a task that we didn't distribute?
    sendNewTask = true // are we going to pop a new task from the queue?

    if syncReq.Caps.NodeId == -1 {
        // new node joining the pool
        nodeId, heartbeat = self.createNode(syncReq.Caps)
    } else {
        // not their first rodeo. there should be a task on the wire.
        taskOnWire = true
        if syncReq.ServerId != self.ServerId {
            // node rejoining rebooted server
            nodeId, heartbeat = self.createNode(syncReq.Caps)
            remembering = true
        } else {
            // supposedly a node reporting back. validate this:
            nodeId = syncReq.Caps.NodeId
            self.nodeLock.Lock()
            heartbeat, ok = self.nodeHeartbeatMap[nodeId]
            self.nodeLock.Unlock()

            if !ok {
                // TODO print a nasty warning
                // probably a node somehow took longer than timeout to report back?
                nodeId, heartbeat = self.createNode(syncReq.Caps)
            } else {
                sendNewTask = false
            }
        }
    }

    // Step 5: Handle task
    if taskOnWire {
        // Step 5.1: Decode task
        err = d.Decode(&oldTask)
        if err != nil {
            http.Error(w, "Could not decode Task", 400)
            return
        }

        // Step 5.2: Process task
        if remembering {
            self.rememberedTasks <- oldTask.Task
        } else {
            self.taskLock.Lock()
            progress, ok := self.taskProgressMap[oldTask.TaskId]
            self.taskLock.Unlock()

            if !ok {
                // task was cancelled
                sendNewTask = true
            } else {
                // if we got this far there was a successful checkpoint
                // let the task watchdog know
                progress <- oldTask.Task
            }
        }
    }

    // Step 6: Pick task to send
    syncResp = SyncResponse{self.Version, self.ServerId, nodeId, "um."}
    if sendNewTask {
        select {
        case newTask = <-self.taskQueue:
            syncResp.Message = "New task!"
        default:
            // no work to do...
            newTask = taskWithId{-1, nil}
            syncResp.Message = "No work to do..."
        }
    } else {
        newTask = taskWithId{oldTask.TaskId, nil}
        syncResp.Message = "Work on old task"
    }

    // Step 7: Update node watchdog with task allocation
    heartbeat <- newTask.TaskId

    // Step 8: Send response!
    buf := bytes.Buffer{}
    e := gob.NewEncoder(&buf)
    err = e.Encode(&syncResp)
    if err != nil {
        http.Error(w, "Could not encode SyncResponse for task..?", 500)
        return
    }

    err = e.Encode(&newTask)
    if err != nil {
        http.Error(w, "Could not encode task", 500)
        return
    }

    _, err = buf.WriteTo(w)
    if err != nil {
        // TODO: log nasty error
    }
}

// one goroutine per node handles the node's membership in the server struct
// timeout watchdog will clean up after the node if it disappears and
// reschedule any dropped jobs
func (self *Server) createNode(caps ClientCaps) (int, chan int) {
    self.nodeEvents <- caps

    heartbeat := make(chan int)

    self.nodeLock.Lock()
    id := self.nextNodeId
    self.nextNodeId++
    self.nodeHeartbeatMap[id] = heartbeat
    self.nodeLock.Unlock()

    go func() {
        curTask := -1
outer:
        for {
            select {
            case curTask = <-heartbeat:
                // keep track of the current task the node is working on
            case <-time.After(self.NodeTimeout):
                // timeout!
                if curTask != -1 {
                    self.taskLock.Lock()
                    progress, ok := self.taskProgressMap[curTask]
                    self.taskLock.Unlock()

                    if ok {
                        // signal to the checkpointer to reschedule the job
                        progress <- nil
                    }   // otherwise the job has been cancelled so no harm done
                }
                break outer
            }
        }

        self.nodeLock.Lock()
        delete(self.nodeHeartbeatMap, id)
        self.nodeLock.Unlock()
    }()

    return id, heartbeat
}

// This is the public method to submit a task
// one goroutine per task handles the task's membership in the server struct
// and forward checkpoints to the output channel
// if block is true it'll block until some node has taken the task
// otherwise we'll return immediately
// the task channel will yield progressive results
// the bool channel can be used to cancel the task
func (self *Server) SubmitTask(t Task, block bool) (chan Task, chan bool) {
    checkpoints := make(chan Task)
    cancel := make(chan bool)

    taskProgress := make(chan Task)

    self.taskLock.Lock()
    id := self.nextTaskId
    self.nextTaskId++
    self.taskProgressMap[id] = taskProgress
    self.taskLock.Unlock()

    if block {
        self.taskQueue <- taskWithId{id, t}
    }

    go func() {
        if !block {
            select {
            case self.taskQueue <- taskWithId{id, t}:
            case <-cancel:
                return
            }
        }

outer:
        for {
            checkpoint := t

            select {
            case progress := <-taskProgress:
                if progress == nil {
                    // node died. resubmit from checkpoint
                    self.taskQueue <- taskWithId{id, checkpoint}
                    continue outer
                }
                checkpoints <- progress
                checkpoint = progress

                if progress.IsDone() {
                    break outer
                }
            case <-cancel:
                // server should check to see if we've deleted the entry from
                // the map to detect cancellation
                break outer
            }
        }

        close(checkpoints)
        self.taskLock.Lock()
        delete(self.taskProgressMap, id)
        self.taskLock.Unlock()
    }()

    return checkpoints, cancel
}
