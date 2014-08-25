package main

import (
    "encoding/json"
    "errors"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "strconv"
)

func getQuestion(db *DB, params ...string) ([]byte, error) {
    if len(params) != 1 {
        return nil, errors.New("Expecting 1 parameter")
    }
    if !bson.IsObjectIdHex(params[0]) {
        return nil, errors.New("Parameter is an invalid BSON ObjectId")
    }
    id := bson.ObjectIdHex(params[0])
    pipe := db.Questions.Pipe([]bson.M{
        {
            "$match": bson.M{
                "_id": id,
            },
        },
        {
            "$unwind": "$revs",
        },
        {
            "$sort": bson.M{
                "revs.ts": 1,
            },
        },
        {
            "$group": bson.M{
                "_id": bson.M{
                    "_id":   "$_id",
                    "ts":    "$ts",
                    "joins": "$joins",
                },
                "last_rev": bson.M{
                    "$last": "$revs",
                },
            },
        },
        {
            "$project": bson.M{
                "_id":     "$_id._id",
                "ts":      "$_id.ts",
                "joins":   "$_id.joins",
                "loc":     "$last_rev.loc",
                "title":   "$last_rev.title",
                "content": "$last_rev.content",
            },
        },
    })
    var q question
    if err := pipe.One(&q); err != nil {
        if err != mgo.ErrNotFound {
            return nil, err
        }
        return nil, nil
    }
    return json.Marshal(q)
}

func getQuestionJoins(db *DB, params ...string) ([]byte, error) {
    if len(params) != 1 {
        return nil, errors.New("Expecting 3 parameters")
    }
    if !bson.IsObjectIdHex(params[0]) {
        return nil, errors.New("First parameter is an invalid BSON ObjectId")
    }
    id := bson.ObjectIdHex(params[0])
    count, err := strconv.Atoi(s)
    if err != nil {
        return nil, errors.New("Second parameter is not an integer")
    }
    page, err := strconv.Atoi(s)
    if err != nil {
        return nil, errors.New("Third parameter is not an integer")
    }
    //TODO
}
