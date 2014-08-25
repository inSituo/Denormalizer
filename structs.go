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
    TS      int      `bson:"ts" json:"ts"`
    Loc     location `bson:"loc" json:"loc"`
    Title   string   `bson:"title" json:"title"`
    Content string   `bson:"content" json:"content"`
    Joins   int      `bson:"joins" json:"joins"`
}

type questionJoin struct {
    Uid  bson.ObjectId `bson:"uid" json:"uid"`
    Disp string        `bson:"disp" json:"disp"`
}

type questionJoins struct {
    Joins []questionJoin `bson:"joins" json:"joins"`
}
