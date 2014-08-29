package main

import (
    "fmt"
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
)

type Server struct {
    port    int
    log     *LeveledLogger.Logger
    db      *DB
    wn      int
    workers []*Worker
    queues  map[int]chan *Work
    wbuff   int
}

func NewServer(port int, wn int, wbuff int, db *DB, log *LeveledLogger.Logger) *Server {
    return &Server{
        port:    port,
        log:     log,
        db:      db,
        wn:      wn,
        workers: make([]*Worker, 0, wn),
        queues:  make(map[int]chan *Work),
        wbuff:   wbuff,
    }
}

func (s *Server) Run() error {
    iname := "Server.Run"
    addr := fmt.Sprintf("tcp://*:%d", s.port)

    // replies queue
    prodq := make(chan *Product, s.wbuff*s.wn)

    s.log.Debug(iname, "creating frontend socket")
    frontend, err := zmq.NewSocket(zmq.ROUTER)
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

    // pool of worker goroutines
    s.log.Debug(iname, "creating workers pool")
    for i := 0; i < s.wn; i++ {
        s.queues[i] = make(chan *Work, s.wbuff)
        worker := NewWorker(i, s.queues[i], prodq, s.db, s.log)
        s.workers = append(s.workers, worker)
        go worker.Run()
        defer worker.Stop()
    }

    // reply to requests pending in the replies quque
    go func() {
        s.log.Debug(iname, "waiting for payloads")
        for {
            prod := <-prodq
            s.log.Debug(iname, "sending reply", prod)
            if _, err := frontend.SendMessage(prod.id, prod.success, prod.empty, prod.payload); err != nil {
                s.log.Warn(iname, "unable to send reply", err)
            }
        }
    }()

    s.log.Info(iname, "listening to incoming requests", addr)
    for {
        if msg, err := frontend.RecvMessage(0); err == nil {
            s.log.Debug(iname, "message received", msg)
            if len(msg) < 3 {
                s.log.Debug(iname, "not enough message parts", len(msg))
                if len(msg) == 2 {
                    prodq <- &Product{
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
            s.assign(&work)
        } else {
            s.log.Warn(iname, "failed to receive incoming message", err)
        }
    }

    return nil

    // now the deferred worker.Stop and frontend.Close will be called
}

func (s *Server) assign(work *Work) {
    iname := "Server.assign"
    i := 0
    l := len(s.queues[i])
    for k, v := range s.queues {
        if len(v) < l {
            i = k
            l = len(v)
        }
    }
    s.log.Debug(iname, "assigning to worker", i)
    s.queues[i] <- work
}
