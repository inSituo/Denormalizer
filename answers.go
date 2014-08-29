package main

import (
    "errors"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

// getAnswer generates a denormalized answer data. It queries the database
// for the data needed to display the answer.
// On success, it returns a pointer to a 'Answer' struct.
// Params: id - The requested answer ID
// Return:
//  1. Pointer to a Answer struct
//  2. (bool) Does the requested answer exist?
//  3. (error) Nil or an error
func (w *Worker) getAnswer(id bson.ObjectId) (*Answer, bool, error) {
    pipe := w.db.Answers.Pipe([]bson.M{
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
                    "qid":        "$qid",
                    "ts":         "$ts",
                    "ranking":    "$ranking",
                    "anon":       "$anon",
                    "thanks":     bson.M{"$size": "$thnksuids"},
                    "thumbups":   bson.M{"$size": "$thupsuids"},
                    "thumbdowns": bson.M{"$size": "$thdwnuids"},
                },
                "first_rev": bson.M{
                    "$first": "$revs",
                },
                "last_rev": bson.M{
                    "$last": "$revs",
                },
            },
        },
        {
            "$project": bson.M{
                "_id":        "$_id._id",
                "qid":        "$_id.qid",
                "lts":        "$_id.ts",
                "ranking":    "$_id.ranking",
                "thanks":     "$_id.thanks",
                "thumbups":   "$_id.thumbups",
                "thumbdowns": "$_id.thumbdowns",
                "anon":       "$_id.anon",
                "fuid":       "$first_rev.uid",
                "fts":        "$first_rev.ts",
                "luid":       "$last_rev.uid",
                "locs":       "$last_rev.locs",
                "content":    "$last_rev.content",
            },
        },
    })
    var a Answer
    if err := pipe.One(&a); err != nil {
        if err != mgo.ErrNotFound {
            return nil, false, err
        }
        return nil, false, nil
    }
    uids := []bson.ObjectId{a.Fuid}
    if a.Fuid != a.Luid {
        uids = append(uids, a.Luid)
    }
    // get the last and first users display data from the db
    query := w.db.Users.
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
    }
    if len(users) < len(uids) {
        return nil, false, errors.New("Unable to find user")
    }
    names := make(map[bson.ObjectId]string)
    for _, v := range users {
        names[v.ID] = v.Name
    }
    // fill-in the missing information in the answer:
    a.Fudisp = names[a.Fuid]
    a.Ludisp = names[a.Luid]
    return &a, true, nil
}
