package main

import (
    "errors"
    "gopkg.in/mgo.v2/bson"
    "strconv"
)

// params parsers:

func parseOid(params []string) (bson.ObjectId, error) {
    if len(params) != 2 {
        return bson.NewObjectId(), errors.New("Incorrect number of arguments")
    }
    id := params[1]
    if !bson.IsObjectIdHex(id) {
        return bson.NewObjectId(), errors.New("Parameter is an invalid BSON ObjectId")
    }
    return bson.ObjectIdHex(id), nil
}

func parseOidCountPage(params []string) (bson.ObjectId, int, int, error) {
    if len(params) != 4 {
        return bson.NewObjectId(), -1, -1, errors.New("Incorrect number of arguments")
    }
    oid := params[1]
    if !bson.IsObjectIdHex(oid) {
        return bson.NewObjectId(), -1, -1, errors.New("First argument is an invalid BSON ObjectId")
    }
    count, err := strconv.Atoi(params[2])
    if err != nil {
        return bson.NewObjectId(), -1, -1, errors.New("Second argument is not an integer")
    }
    page, err := strconv.Atoi(params[3])
    if err != nil {
        return bson.NewObjectId(), -1, -1, errors.New("Third argument is not an integer")
    }
    return bson.ObjectIdHex(oid), count, page, nil
}
