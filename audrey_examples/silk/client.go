package silk

import (
    "fmt"
    "time"
    "bytes"
    "io/ioutil"
    "encoding/gob"
)

type clientError struct {
    message string
    orig error
}

func (self clientError) Error() string {
    if self.orig == nil {
        return self.message
    } else {
        return fmt.Sprintf("%s: %s", self.message, self.orig.Error())
    }
}

func (self *Client) Run() (int, error) {
    if self.running {
        panic("Called Run() on already running client!")
    }
    self.running = true
    defer func(){self.running = false}()

    if self.Caps.NodeId == 0 {
        self.Caps.NodeId = -1
    }

    self.serverId = -1

    t, v, err := self.sync(nil)
    if err != nil {
        return v, err
    }

    cancel := make(chan bool)
    progress := make(chan Task)
    go t.Run(progress, cancel)

    for {
        select {
        case newt := <-progress:
            t, v, err = self.sync(newt)
        case <-time.After(time.Duration(30 * time.Second)):
            t, v, err = self.sync(nil)
        }

        if err != nil {
            cancel <- true
            return v, err
        }

        if t != nil {
            cancel <- true
            cancel = make(chan bool)
            progress = make(chan Task)
            go t.Run(progress, cancel)
        }
    }
}

func (self *Client) sync(oldTask Task) (Task, int, error) {
    var sync SyncResponse
    var err error

    buf := bytes.Buffer{}
    e := gob.NewEncoder(&buf)

    err = e.Encode(&SyncRequest{self.Version, self.serverId, self.Caps})
    if err != nil {
        return nil, 0, clientError{"Couldn't encode SyncRequest", err}
    }

    if oldTask != nil || self.Caps.NodeId != -1 {
        err = e.Encode(&oldTask)
        if err != nil {
            return nil, 0, clientError{"Couldn't encode task", err}
        }
    }

    url := fmt.Sprintf("http://%s:%d/sync", self.ServerDomain, self.ServerPort)
    resp, err := self.netClient.Post(url, "application/octet-stream", &buf)
    if err != nil {
        return nil, 0, clientError{"Sync transport failed", err}
    }

    if resp.StatusCode >= 400 {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return nil, 0, clientError{fmt.Sprintf("Sync failed with HTTP %d: [body undecodable: %s]", resp.StatusCode, err.Error()), nil}
        } else {
            return nil, 0, clientError{fmt.Sprintf("Sync failed with HTTP %d: %s", resp.StatusCode, string(body)), nil}
        }
    }

    d := gob.NewDecoder(resp.Body)
    err = d.Decode(&sync)
    if err != nil {
        return nil, 0, clientError{"Couldn't decode SyncResponse", err}
    }

    if sync.Version != self.Version {
        return nil, sync.Version, clientError{"Must upgrade!", nil}
    }

    self.serverId = sync.ServerId
    self.Caps.NodeId = sync.NodeId

    var newTask Task
    err = d.Decode(&newTask)
    if err != nil {
        return nil, 0, clientError{"Couldn't decode response task", err}
    }

    resp.Body.Close()
    return newTask, 0, nil
}
