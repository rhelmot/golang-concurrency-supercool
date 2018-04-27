package silk

import (
    "sync"
    "time"
    "net/http"
    "encoding/gob"
)

type Task interface {
    IsDone() bool
    Run(chan Task, chan bool)
}

type ClientCaps struct {
    NodeId int
    CapSites int
    CapMemMB int
    CapCpus int
    CapLifetime time.Duration
}

type SyncRequest struct {
    Version int
    ServerId int
    Caps ClientCaps
}

type SyncResponse struct {
    Version int
    ServerId int
    NodeId int
    Message string
}

type Server struct {
    Version int
    Listen string
    NodeTimeout time.Duration
    ServerId int

    serving bool

    taskQueue chan taskWithId
    rememberedTasks chan Task
    nodeEvents chan ClientCaps

    taskLock sync.Mutex
    taskProgressMap map[int]chan Task
    nextTaskId int

    nodeLock sync.Mutex
    nodeHeartbeatMap map[int]chan int
    nextNodeId int
}

type Client struct {
    Version int
    ServerDomain string
    ServerPort int
    Caps ClientCaps

    running bool

    serverId int
    netClient http.Client
}

func RegisterTaskType(value Task) {
    gob.Register(value)
}
