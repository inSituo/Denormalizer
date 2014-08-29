package main

import (
    "gopkg.in/mgo.v2/bson"
)

type Coordinates struct {
    Lat float32 `bson:"lat" json:"lat"`
    Lon float32 `bson:"lon" json:"lon"`
}

type Location struct {
    Crd  Coordinates `bson:"crd" json:"crd"`
    Path []string    `bson:"path" json:"path"`
}

type Question struct {
    ID      bson.ObjectId `bson:"_id" json:"id"`
    TS      int           `bson:"ts" json:"ts"`
    Loc     Location      `bson:"loc" json:"loc"`
    Title   string        `bson:"title" json:"title"`
    Content string        `bson:"content" json:"content"`
    Joins   int           `bson:"joins" json:"joins"`
}

type QuestionJoin struct {
    Uid   bson.ObjectId `bson:"uid" json:"uid"`
    Udisp string        `bson:"udisp" json:"udisp"`
}

type Comment struct {
    ID      bson.ObjectId `bson:"_id" json:"id"`
    Uid     bson.ObjectId `bson:"uid" json:"uid"`
    Udisp   string        `bson:"udisp" json:"udisp"`
    TS      int           `bson:"ts" json:"ts"`
    Content string        `bson:"content" json:"content"`
}

type Answer struct {
    ID         bson.ObjectId `bson:"_id" json:"id"`
    LTS        int           `bson:"lts" json:"lts"`
    FTS        int           `bson:"fts" json:"fts"`
    Qid        bson.ObjectId `bson:"qid" json:"qid"`
    Fuid       bson.ObjectId `bson:"fuid" json:"fuid"`
    Luid       bson.ObjectId `bson:"luid" json:"luid"`
    Fudisp     string        `bson:"fudisp" json:"fudisp"`
    Ludisp     string        `bson:"ludisp" json:"ludisp"`
    Anon       bool          `bson:"anon" json:"anon"`
    Locs       []Location    `bson:"locs" json:"locs"`
    Content    string        `bson:"content" json:"content"`
    Ranking    int           `bson:"ranking" json:"ranking"`
    Thanks     int           `bson:"thanks" json:"thanks"`
    Thumbups   int           `bson:"thumbups" json:"thumbups"`
    Thumbdowns int           `bson:"thumbdowns" json:"thumbdowns"`
}
