package main

import (
    "fmt"
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
    "os"
    // "runtime"
    "sync"
    "time"
)

type Server struct {
    port     int
    log      *LeveledLogger.Logger
    ll_level int
    dbconf   *MongoConf
    wn       int
    workers  []*Worker
    queues   map[int]chan *Work
    wbuff    int
}

func NewServer(port int, wn int, wbuff int, dbconf *MongoConf, ll_level int) *Server {
    return &Server{
        port:     port,
        log:      LeveledLogger.New(os.Stdout, ll_level),
        dbconf:   dbconf,
        wn:       wn,
        workers:  make([]*Worker, 0, wn),
        queues:   make(map[int]chan *Work),
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
        s.dbconf.Host,
        s.dbconf.Port,
        s.dbconf.DB,
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

    // replies queue
    outgoing := make(chan *Product, s.wbuff*s.wn)

    // pool of worker goroutines
    s.log.Debug(iname, "creating workers pool")
    for i := 0; i < s.wn; i++ {
        s.queues[i] = make(chan *Work, s.wbuff)
        worker := NewWorker(i, s.queues[i], outgoing, db, s.ll_level)
        s.workers = append(s.workers, worker)
        go worker.Run()
        defer worker.Stop()
    }

    incoming := make(chan []string, s.wbuff*s.wn)

    // receiver:
    go func() {
        s.log.Info(iname, "listening to incoming requests", addr)
        for {
            felock.Lock() // lock socket access
            msg, err := frontend.RecvMessage(zmq.DONTWAIT)
            felock.Unlock() // release socket access
            if err == nil {
                incoming <- msg
            } else {
                s.log.Warn(iname, "failed to receive incoming message", err)
            }
            // receiving 500 requests/second should be enough and not kill
            // the cpu.
            time.Sleep(2 * time.Millisecond)
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
            s.assign(&work)
        }
    }()

    // reply to requests pending in the replies quque
    s.log.Debug(iname, "waiting for payloads")
    for {
        prod := <-outgoing
        s.log.Debug(iname, "sending reply", prod)
        felock.Lock() // lock socket access
        _, err := frontend.SendMessage(prod.id, prod.success, prod.empty, prod.payload)
        felock.Unlock() // release socket access
        if err != nil {
            s.log.Warn(iname, "unable to send reply", err)
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
