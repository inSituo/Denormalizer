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
    if len(params) != 3 {
        return nil, errors.New("Expecting 3 parameters")
    }
    if !bson.IsObjectIdHex(params[0]) {
        return nil, errors.New("First parameter is an invalid BSON ObjectId")
    }
    id := bson.ObjectIdHex(params[0])
    count, err := strconv.Atoi(params[1])
    if err != nil {
        return nil, errors.New("Second parameter is not an integer")
    }
    page, err := strconv.Atoi(params[2])
    if err != nil {
        return nil, errors.New("Third parameter is not an integer")
    }
    pipe := db.Questions.Pipe([]bson.M{
        {
            "$match": bson.M{
                "_id": id,
            },
        },
        {
            "$unwind": "$juids",
        },
        {
            "$skip": count * page,
        },
        {
            "$limit": count,
        },
        {
            "$group": bson.M{
                "_id":   bson.M{"_id": "$_id"},
                "juids": bson.M{"$push": "$juids"},
            },
        },
        {
            "$project": bson.M{
                "_id":   false,
                "juids": "$juids",
            },
        },
    })
    var juids map[string][]bson.ObjectId
    if err := pipe.One(&juids); err != nil {
        if err != mgo.ErrNotFound {
            return nil, err
        }
        return nil, nil
    }
    qjs := make([]questionJoin, 0, count)
    pipe = db.Users.Pipe([]bson.M{
        {
            "$match": bson.M{
                "_id": bson.M{"$in": juids["juids"]},
            },
        },
        {
            "$project": bson.M{
                "_id":   false,
                "uid":   "$_id",
                "udisp": "$name",
            },
        },
    })
    if err := pipe.All(&qjs); err != nil {
        if err != mgo.ErrNotFound {
            return nil, err
        }
        return nil, nil
    }
    return json.Marshal(questionJoins{Joins: qjs})
}

func getQuestionLatestComments(db *DB, params ...string) ([]byte, error) {
    if len(params) != 3 {
        return nil, errors.New("Expecting 3 parameters")
    }
    if !bson.IsObjectIdHex(params[0]) {
        return nil, errors.New("First parameter is an invalid BSON ObjectId")
    }
    id := bson.ObjectIdHex(params[0])
    count, err := strconv.Atoi(params[1])
    if err != nil {
        return nil, errors.New("Second parameter is not an integer")
    }
    page, err := strconv.Atoi(params[2])
    if err != nil {
        return nil, errors.New("Third parameter is not an integer")
    }
    cmts := make([]comment, 0)
    query := db.Comments.
        Find(bson.M{"oid": id, "type": "question"}).
        Sort("-ts").
        Skip(count * page).
        Limit(count).
        Select(bson.M{
        "_id":     true,
        "uid":     true,
        "ts":      true,
        "content": true})
    if err := query.All(&cmts); err != nil {
        if err != mgo.ErrNotFound {
            return nil, err
        }
        return nil, nil
    }
    uids := make([]bson.ObjectId, 0, len(cmts))
    for _, v := range cmts {
        uids = append(uids, v.Uid)
    }
    query = db.Users.
        Find(bson.M{"_id": bson.M{"$in": uids}}).
        Select(bson.M{"_id": true, "name": true})
    users := make([]struct {
        ID   bson.ObjectId `bson:"_id"`
        Name string        `bson:"name"`
    }, len(uids))
    if err := query.All(&users); err != nil {
        if err != mgo.ErrNotFound {
            return nil, err
        }
        return nil, nil
    }
    names := make(map[bson.ObjectId]string)
    for _, v := range users {
        names[v.ID] = v.Name
    }
    for i, _ := range cmts {
        cmts[i].Udisp = names[cmts[i].Uid]
    }
    return json.Marshal(comments{Comments: cmts})
}
