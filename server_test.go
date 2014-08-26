package main

import (
    "github.com/inSituo/LeveledLogger"
    zmq "github.com/pebbe/zmq4"
    "os"
    "testing"
    "time"
)

func TestServer(t *testing.T) {
    iname := "ServerTest"

    port := 27017
    host := "127.0.0.1"
    dbname := "insituo-tst"
    cusers := "users"
    cquestions := "questions"
    canswers := "answers"

    log := LeveledLogger.New(os.Stdout, LeveledLogger.LL_DEBUG)
    db, err := NewDB(MongoConf{
        Port:       &port,
        Host:       &host,
        DB:         &dbname,
        CUsers:     &cusers,
        CQuestions: &cquestions,
        CAnswers:   &canswers,
    })
    if err != nil {
        log.Error(iname, "failed to create a new db", err) // this will panic
    }
    defer db.Close()

    server, err := NewServer(1234, log, db)
    if err != nil {
        log.Error(iname, "failed to create server", err) // this will panic
    }

    go func() {
        err = server.Run(10)
        log.Error(iname, "server stopped", err) // this will panic
    }()
    time.Sleep(1000 * time.Millisecond)
    return
    // simulate clients load - 15 clients every 100ms
    for {
        for i := 0; i < 15; i++ {
            go func() {
                client, err := zmq.NewSocket(zmq.REQ)
                if err != nil {
                    t.Error(err)
                    t.Fail()
                    return
                }
                defer client.Close()

                client.Connect("tcp://127.0.0.1:1234")

                client.SendMessage("questionJoins", "53fb63a4472dcb6b32e99260", "5", "1")

                _, err = client.RecvMessage(0)
                if err != nil {
                    t.Error(err)
                    t.Fail()
                }
            }()
        }
        time.Sleep(100 * time.Millisecond)
    }

}
