package main

import (
    "gopkg.in/mgo.v2/bson"
)

type coordinates struct {
    Lat float32 `bson:"lat" json:"lat"`
    Lon float32 `bson:"lon" json:"lon"`
}

type location struct {
    Crd  coordinates `bson:"crd" json:"crd"`
    Path []string    `bson:"path" json:"path"`
}

type question struct {
    ID      bson.ObjectId `bson:"_id" json:"id"`
    TS      int           `bson:"ts" json:"ts"`
    Loc     location      `bson:"loc" json:"loc"`
    Title   string        `bson:"title" json:"title"`
    Content string        `bson:"content" json:"content"`
    Joins   int           `bson:"joins" json:"joins"`
}

type questionJoin struct {
    Uid   bson.ObjectId `bson:"uid" json:"uid"`
    Udisp string        `bson:"udisp" json:"udisp"`
}

type questionJoins struct {
    Joins []questionJoin `bson:"joins" json:"joins"`
}

type comment struct {
    ID      bson.ObjectId `bson:"_id" json:"id"`
    Uid     bson.ObjectId `bson:"uid" json:"uid"`
    Udisp   string        `bson:"udisp" json:"udisp"`
    TS      int           `bson:"ts" json:"ts"`
    Content string        `bson:"content" json:"content"`
}

type comments struct {
    Comments []comment `bson:"comments" json:"comments"`
}

type answer struct {
    ID         bson.ObjectId `bson:"_id" json:"id"`
    TS         int           `bson:"ts" json:"ts"`
    Uid        bson.ObjectId `bson:"uid" json:"uid"`
    Locs       []location    `bson:"locs" json:"locs"`
    Content    string        `bson:"content" json:"content"`
    Ranking    int           `bson:"ranking" json:"ranking"`
    Thanks     int           `bson:"ranking" json:"ranking"`
    Thumbups   int           `bson:"ranking" json:"ranking"`
    Thumbdowns int           `bson:"ranking" json:"ranking"`
}
