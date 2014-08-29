package main

import (
    "flag"
    "fmt"
    "github.com/inSituo/LeveledLogger"
    "os"
)

type DenormConf struct {
    mongo   MongoConf
    port    *int
    workers *int
    wbuff   *int
    debug   *bool
}

func main() {
    iname := "main"
    conf := DenormConf{
        debug:   flag.Bool("debug", false, "Enable debug log messages"),
        port:    flag.Int("port", 7710, "ZeroMQ listening port"),
        workers: flag.Int("workers", 5, "Number of workers"),
        wbuff:   flag.Int("buffer", 100, "Size of one worker's buffer"),
        mongo: MongoConf{
            Port:       flag.Int("mport", 27017, "MongoDB server port"),
            Host:       flag.String("mhost", "127.0.0.1", "MongoDB server host"),
            DB:         flag.String("mdb", "insituo-dev", "MongoDB database name"),
            CUsers:     flag.String("cusers", "users", "Name of users collection in DB"),
            CQuestions: flag.String("cquestions", "questions", "Name of questions collection in DB"),
            CAnswers:   flag.String("canswers", "answers", "Name of answers collection in DB"),
            CComments:  flag.String("ccomments", "comments", "Name of comments collection in DB"),
        },
    }

    showHelp := flag.Bool("help", false, "Show help")

    flag.Parse()
    if *showHelp {
        fmt.Printf("Denormalizer %s\n\n",
            fmt.Sprintf(
                "v%d.%d.%d",
                DENORM_VER_MAJOR,
                DENORM_VER_MINOR,
                DENORM_VER_PATCH,
            ),
        )
        flag.PrintDefaults()
        return
    }

    ll_level := LeveledLogger.LL_INFO
    if *conf.debug {
        ll_level = LeveledLogger.LL_DEBUG
    }
    log := LeveledLogger.New(os.Stdout, ll_level)

    log.Debug(iname, "debug mode enabled")

    log.Info(
        iname,
        "connecting to MongoDB",
        *conf.mongo.Host,
        *conf.mongo.Port,
        *conf.mongo.DB,
    )
    db, err := NewDB(conf.mongo)
    if err != nil {
        log.Error(iname, "failed to create a new db", err) // this will panic
    }
    defer db.Close()

    server := NewServer(*conf.port, *conf.workers, *conf.wbuff, db, log)
    err = server.Run()
    log.Error(iname, "server run error", err) // this will panic

}
