package main

import (
    "errors"
    "fmt"
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
)

func worker(num int, db *DB, log *LeveledLogger.Logger) {
    iname := fmt.Sprintf("worker %d", num)

    worker, err := zmq.NewSocket(zmq.REP)
    if err != nil {
        log.Warn(iname, "new socket failed", err)
        return
    }
    defer worker.Close()
    worker.Connect(WORKERS_COMM_CHANNEL)

    log.Debug(iname, "ready")

    for {
        if msg, err := worker.RecvMessage(0); err == nil {

            imsg := make([]interface{}, len(msg))
            for i, v := range msg {
                imsg[i] = interface{}(v)
            }
            log.Info(iname, "message received", imsg...)

            var res []byte
            var err error

            switch msg[0] {
            case "question":
                res, err = getQuestion(db, msg[1:]...)
            case "questionJoins":
                res, err = getQuestionJoins(db, msg[1:]...)
            case "ping":
                res = []byte("pong")
            default:
                err = errors.New(fmt.Sprintf("unknown task: %s", msg[0]))
            }

            if err == nil {
                log.Info(iname, "task completed", msg[0])
                if _, err := worker.SendMessage("result", res); err != nil {
                    log.Warn(iname, "message send failure", err)
                }
            } else {
                log.Warn(iname, "task failed", err)
                if _, err := worker.SendMessage("error", err); err != nil {
                    log.Warn(iname, "error message send failure", err)
                }
            }

        } else {
            log.Warn(iname, "message receiving error", err)
        }
    }
}