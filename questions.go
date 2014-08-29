package main

import (
    "errors"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// getQuestion generates a denormalized question data. It queries the database
// for the data needed to display the question.
// On success, it returns a pointer to a 'Question' struct.
// Params: id - The requested question ID
// Return:
//  1. Pointer to a Question struct
//  2. (bool) Does the requested question exist?
//  3. (error) Nil or an error
func (w *Worker) getQuestion(id bson.ObjectId) (*Question, bool, error) {
    // we use aggregation to bring question to its denormalized form.
    // we only need the last revision of the question's content.
    pipe := w.db.Questions.Pipe([]bson.M{
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
    var q Question
    if err := pipe.One(&q); err != nil {
        if err != mgo.ErrNotFound {
            return nil, false, err
        }
        // aggregation returned empty
        return nil, false, nil
    }
    return &q, true, nil
}

// getQuestionJoins generates a denormalized question joins data.
// The result is an array of user ids and user names, of all the users that
// joined the question.
// On success, it returns a pointer to a 'QuestionJoins' struct.
// Params:
//  1. id - The requested question ID
//  2. count - how many joins to return
//  3. page - page offset of joins array (starting from 0)
// Return:
//  1. Pointer to a QuestionJoins struct
//  2. (bool) Does the requested question exist?
//  3. (error) Nil or an error
func (w *Worker) getQuestionJoins(id bson.ObjectId, count, page int) (*QuestionJoins, bool, error) {
    pipe := w.db.Questions.Pipe([]bson.M{
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
    // the result is a one element map:
    // { "juids": [id1, id2, id3, ...]}
    var juids map[string][]bson.ObjectId
    if err := pipe.One(&juids); err != nil {
        if err != mgo.ErrNotFound {
            return nil, false, err
        }
        return nil, false, nil
    }
    qjs := make([]QuestionJoin, 0, count)
    pipe = w.db.Users.Pipe([]bson.M{
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
        return nil, false, err
    }
    return &QuestionJoins{Joins: qjs}, true, nil
}

// getQuestionLatestComments generates a denormalized question comments data.
// The result is an array of user comments (user id, user name, time, content)
// On success, it returns a pointer to a 'Comments' struct.
// Params:
//  1. id - The requested question ID
//  2. count - how many comments to return
//  3. page - page offset of comments array (starting from 0)
// Return:
//  1. Pointer to a Comments struct
//  2. (bool) Does the requested question exist?
//  3. (error) Nil or an error
func (w *Worker) getQuestionLatestComments(id bson.ObjectId, count, page int) (*Comments, bool, error) {
    cmts := make([]Comment, 0)
    query := w.db.Comments.
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
            return nil, false, err
        }
        return nil, false, nil
    }
    uids := make([]bson.ObjectId, 0, len(cmts))
    for _, v := range cmts {
        uids = append(uids, v.Uid)
    }
    query = w.db.Users.
        Find(bson.M{"_id": bson.M{"$in": uids}}).
        Select(bson.M{"_id": true, "name": true})
    users := make([]struct {
        ID   bson.ObjectId `bson:"_id"`
        Name string        `bson:"name"`
    }, len(uids))
    if err := query.All(&users); err != nil {
        if err != mgo.ErrNotFound {
            return nil, false, err
        }
        return nil, false, errors.New("Unable to find user")
    }
    names := make(map[bson.ObjectId]string)
    for _, v := range users {
        names[v.ID] = v.Name
    }
    for i, _ := range cmts {
        cmts[i].Udisp = names[cmts[i].Uid]
    }
    return &Comments{Comments: cmts}, true, nil
}
