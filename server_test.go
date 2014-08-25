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
        err = server.Run(5)
        log.Error(iname, "server interrupted", err) // this will panic
    }()

    // simulate 10 clients:
    for i := 0; i < 10; i++ {
        go func(i int) {
            client, _ := zmq.NewSocket(zmq.REQ)
            defer client.Close()

            client.Connect("tcp://127.0.0.1:1234")

            client.SendMessage("ping")

            msg, err := client.RecvMessage(0)
            if err == nil {
                if msg[1] != "pong" {
                    t.Errorf("Expected response 'ping' to 'pong'. Received %s", msg[1])
                }
            } else {
                t.Error(err)
            }
        }(i)

    }

    //test 'questons'
    go func() {
        client, _ := zmq.NewSocket(zmq.REQ)
        defer client.Close()

        client.Connect("tcp://127.0.0.1:1234")

        client.SendMessage("question", "53fb63a4472dcb6b32e99260")

        msg, err := client.RecvMessage(0)
        if err == nil {
            if msg[1] != "pong" {
                t.Errorf("Expected response 'ping' to 'pong'. Received %s", msg[1])
            }
        } else {
            t.Error(err)
        }
    }()

    time.Sleep(300 * time.Millisecond)
}
