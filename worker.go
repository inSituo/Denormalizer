package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/inSituo/LeveledLogger"
    "gopkg.in/mgo.v2/bson"
    "os"
)

// Work to be done by a worker
type Work struct {
    // The ID of the message received by the socket.
    // Will be needed by the router to know to which client to send the
    // response.
    id  []string

    // The rest of the message parts as received from the client request.
    // Used to determine the task to be performed by the workers and the
    // arguments for this task.
    params []string
}

// The result of a worker's work.
type Product struct {
    // The ID corresponding to the work which produced this product.
    // Will be needed by the router to know to which client to send the
    // response.
    id  []string

    // Did the work succeed? Or are there errors? If there is an error, the
    // description is in the 'payload' field.
    success bool

    // Even if the work succeeded, is there any payload? Maybe the result is
    // empty.
    empty bool

    // The result of the work, or an error description.
    // If the work succeeded, this is the result encoded as a JSON string.
    payload []byte
}

// A worker receives work through a 'work queue' and produces products. The
// produced products are queued in a 'products queue'.
type Worker struct {
    // The ID of the worker, which is assigned to it when created.
    ID  int

    db  *DB

    log *LeveledLogger.Logger

    // A buffered channel which the worker constantly polls for new work.
    workq chan *Work

    // A buffered channel into which the worker pushes work products.
    prodq chan *Product

    // An unbuffered channel which is used to signal the worker to shut down.
    stopc chan bool
}

// Construct a new worker object.
// Notice that the 'db' connection is copied, and not used as is. This means
// a new database connection will be made.
func NewWorker(
    id int,
    workq chan *Work,
    prodq chan *Product,
    db *DB,
    ll_level int,
) *Worker {
    return &Worker{
        ID:    id,
        db:    db.Copy(),
        log:   LeveledLogger.New(os.Stdout, ll_level),
        workq: workq,
        prodq: prodq,
        stopc: make(chan bool),
    }
}

// Shut down the worker.
func (w *Worker) Stop() {
    w.stopc <- true
    w.db.Close()
    <-w.stopc
}

// Run the worker and start polling for new work.
func (w *Worker) Run() {
    iname := fmt.Sprintf("Worker(%d)", w.ID)

    w.log.Debug(iname, "ready")

    for {
        select {
        case work := <-w.workq:
            var err error
            var exists bool
            var res interface{}
            switch work.params[0] {
            case "Q": // question
                var qid bson.ObjectId
                qid, err = parseOid(work.params)
                if err == nil {
                    res, exists, err = w.GetQuestion(qid)
                }
            case "QJ": // question joins
                var count, page int
                var qid bson.ObjectId
                qid, count, page, err = parseOidCountPage(work.params)
                if err == nil {
                    res, exists, err = w.GetQuestionJoins(qid, count, page)
                }
            case "QLC": // question latest comments
                var count, page int
                var qid bson.ObjectId
                qid, count, page, err = parseOidCountPage(work.params)
                if err == nil {
                    res, exists, err = w.GetQuestionLatestComments(qid, count, page)
                }
            case "A": // answer
                var aid bson.ObjectId
                aid, err = parseOid(work.params)
                if err == nil {
                    res, exists, err = w.GetAnswer(aid)
                }
            case "QTA": // question top answers
                var count, page int
                var qid bson.ObjectId
                qid, count, page, err = parseOidCountPage(work.params)
                if err == nil {
                    res, exists, err = w.GetTopAnswers(qid, count, page)
                }
            case "QLA": // question latest answers
                var count, page int
                var qid bson.ObjectId
                qid, count, page, err = parseOidCountPage(work.params)
                if err == nil {
                    res, exists, err = w.GetLatestAnswers(qid, count, page)
                }
            default:
                err = errors.New("unknown task")
            }
            var payload []byte
            if err == nil && exists {
                payload, err = json.Marshal(res)
            }
            if err == nil {
                w.log.Info(iname, "task completed", work.params[0])
                w.prodq <- &Product{
                    id:      work.id,
                    success: true,
                    empty:   !exists,
                    payload: payload,
                }
            } else {
                w.log.Warn(iname, "task failed", work.params[0], err)
                w.prodq <- &Product{
                    id:      work.id,
                    success: false,
                    empty:   false,
                    payload: []byte(err.Error()),
                }
            }
        case <-w.stopc:
            w.log.Debug(iname, "stopped")
            // release the Stop method:
            defer func() { w.stopc <- true }()
            return
        }
    }
}
