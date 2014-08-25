package main

import (
    "fmt"
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
)

const (
    WORKERS_COMM_CHANNEL = "inproc://workerscomm"
)

type Server struct {
    port     int
    log      *LeveledLogger.Logger
    db       *DB
    frontend *zmq.Socket
    backend  *zmq.Socket
}

func NewServer(port int, log *LeveledLogger.Logger, db *DB) (*Server, error) {
    iname := "NewServer"

    s := &Server{
        port: port,
        log:  log,
        db:   db,
    }
    var err error

    log.Debug(iname, "creating frontend socket")
    s.frontend, err = zmq.NewSocket(zmq.ROUTER)
    if err != nil {
        return nil, err
    }
    if err := s.frontend.Bind(
        fmt.Sprintf("tcp://*:%d", port),
    ); err != nil {
        return nil, err
    }

    log.Debug(iname, "creating backend socket")
    s.backend, err = zmq.NewSocket(zmq.DEALER)
    if err != nil {
        return nil, err
    }
    if err := s.backend.Bind(WORKERS_COMM_CHANNEL); err != nil {
        return nil, err
    }

    return s, nil
}

func (s *Server) Run(n int) error {
    iname := "Server.Run"
    defer s.frontend.Close()
    defer s.backend.Close()

    // pool of worker threads
    s.log.Debug(iname, "creating workers pool")
    for i := 0; i < n; i++ {
        go worker(i, s.db, s.log)
    }

    //  Connect workers to clients via a proxy
    s.log.Info(iname, "waiting for connections", fmt.Sprintf("tcp://*:%d", s.port))
    return zmq.Proxy(s.frontend, s.backend, nil)
}
