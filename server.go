package main

import (
    "fmt"
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
    "os"
    "sync"
    "time"
)

const (
    POLL_TIMEOUT  = 1 * time.Millisecond
    THREADS_SLEEP = 5 * time.Millisecond
)

type Server struct {
    port     int
    log      *LeveledLogger.Logger
    ll_level int
    dbconf   *MongoConf
    wn       int
    wbuff    int
}

func NewServer(port int, wn int, wbuff int, dbconf *MongoConf, ll_level int) *Server {
    return &Server{
        port:     port,
        log:      LeveledLogger.New(os.Stdout, ll_level),
        dbconf:   dbconf,
        wn:       wn,
        wbuff:    wbuff,
        ll_level: ll_level,
    }
}

func (s *Server) Run() error {
    iname := "Server.Run"
    addr := fmt.Sprintf("tcp://*:%d", s.port)

    s.log.Info(
        iname,
        "connecting to MongoDB",
        *s.dbconf.Host,
        *s.dbconf.Port,
        *s.dbconf.DB,
    )
    db, err := NewDB(s.dbconf)
    if err != nil {
        return err
    }
    defer db.Close()

    s.log.Debug(iname, "creating frontend socket")
    felock := sync.Mutex{}
    context, err := zmq.NewContext()
    if err != nil {
        return err
    }
    frontend, err := context.NewSocket(zmq.ROUTER)
    if err != nil {
        return err
    }
    defer frontend.Close()
    s.log.Debug(iname, "binding frontend socket", addr)
    if err := frontend.Bind(
        fmt.Sprintf("tcp://*:%d", s.port),
    ); err != nil {
        return err
    }

    outgoing := make(chan *Product, s.wbuff*s.wn)
    incoming := make(chan []string, s.wbuff*s.wn)
    workq := make(chan *Work, s.wbuff*s.wn)

    // pool of worker goroutines
    s.log.Debug(iname, "creating workers pool")
    for i := 0; i < s.wn; i++ {
        worker := NewWorker(i, workq, outgoing, db, s.ll_level)
        go worker.Run()
        defer worker.Stop()
    }

    // receiver:
    go func() {
        poller := zmq.NewPoller()
        poller.Add(frontend, zmq.POLLIN)
        s.log.Info(iname, "listening to incoming requests", addr)
        for {
            felock.Lock()
            polled, err := poller.Poll(POLL_TIMEOUT)
            if err == nil {
                if len(polled) > 0 {
                    msg, err := frontend.RecvMessage(0)
                    if err == nil {
                        incoming <- msg
                    } else {
                        s.log.Warn(iname, "failed to receive incoming message", err)
                    }
                }
            } else {
                s.log.Warn(iname, "failed to poll socket", err)
            }
            felock.Unlock()
            // give other threads a chance to obtain the lock:
            time.Sleep(THREADS_SLEEP)
        }
    }()

    // dispatcher
    go func() {
        for {
            msg := <-incoming
            s.log.Debug(iname, "message received", msg)
            if len(msg) < 3 {
                s.log.Debug(iname, "not enough message parts", len(msg))
                if len(msg) == 2 {
                    outgoing <- &Product{
                        id:      msg,
                        success: false,
                        empty:   false,
                        payload: []byte("no task specified"),
                    }
                }
                continue
            }
            work := Work{
                id:     msg[:2],
                params: msg[2:],
            }
            workq <- &work
        }
    }()

    // reply to requests pending in the replies quque
    s.log.Debug(iname, "waiting for payloads")
    for {
        prod := <-outgoing
        s.log.Debug(iname, "sending reply", prod)
        felock.Lock()
        _, err := frontend.SendMessage(prod.id, prod.success, prod.empty, prod.payload)
        felock.Unlock()
        if err != nil {
            s.log.Warn(iname, "unable to send reply", err)
        }
        // give other threads a chance to obtain the lock:
        time.Sleep(THREADS_SLEEP)
    }
    return nil

    // now the deferred worker.Stop and frontend.Close will be called
}
