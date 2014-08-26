package main

import (
    "encoding/json"
    "errors"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "strconv"
)

func getAnswer(db *DB, params ...string) ([]byte, error) {
    if len(params) != 1 {
        return nil, errors.New("Expecting 1 parameter")
    }
    if !bson.IsObjectIdHex(params[0]) {
        return nil, errors.New("Parameter is an invalid BSON ObjectId")
    }
    id := bson.ObjectIdHex(params[0])
    pipe := db.Answers.Pipe([]bson.M{
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
                    "_id":        "$_id",
                    "ts":         "$ts",
                    "ranking":    "$ranking",
                    "thanks":     bson.M{"$size": "$thnksuids"},
                    "thumbups":   bson.M{"$size": "$thupsuids"},
                    "thumbdowns": bson.M{"$size": "$thdwnuids"},
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
                "ranking": "$_id.ranking",
                "locs":    "$last_rev.locs",
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

func getAnswerLatestComments(db *DB, params ...string) ([]byte, error) {
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

func getQuestionLatestAnswers(db *DB, params ...string) ([]byte, error) {
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
