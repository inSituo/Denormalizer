package main

import (
    "fmt"
    "gopkg.in/mgo.v2"
)

type MongoConf struct {
    Port       *int
    Host       *string
    DB         *string
    CUsers     *string
    CQuestions *string
    CAnswers   *string
}

type DB struct {
    Users     *mgo.Collection
    Questions *mgo.Collection
    Answers   *mgo.Collection
    session   *mgo.Session
}

func NewDB(conf MongoConf) (*DB, error) {
    url := fmt.Sprintf("%s:%d", *conf.Host, *conf.Port)

    msession, err := mgo.Dial(url)
    if err != nil {
        return nil, err
    }

    mdb := msession.DB(*conf.DB)
    db := &DB{
        Users:     mdb.C(*conf.CUsers),
        Questions: mdb.C(*conf.CQuestions),
        Answers:   mdb.C(*conf.CAnswers),
        session:   msession,
    }

    return db, nil
}

func (db *DB) Close() {
    db.session.Close()
}
