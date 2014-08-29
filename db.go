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
    CComments  *string
}

type DB struct {
    Users     *mgo.Collection
    Questions *mgo.Collection
    Answers   *mgo.Collection
    Comments  *mgo.Collection
    session   *mgo.Session
    conf      *MongoConf
}

func NewDB(conf *MongoConf) (*DB, error) {
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
        Comments:  mdb.C(*conf.CComments),
        session:   msession,
        conf:      conf,
    }

    return db, nil
}

func (db *DB) Copy() *DB {
    s := db.session.Copy()
    mdb := s.DB(*db.conf.DB)
    return &DB{
        Users:     mdb.C(*db.conf.CUsers),
        Questions: mdb.C(*db.conf.CQuestions),
        Answers:   mdb.C(*db.conf.CAnswers),
        Comments:  mdb.C(*db.conf.CComments),
        session:   s,
        conf:      db.conf,
    }
}

func (db *DB) Close() {
    db.session.Close()
}
