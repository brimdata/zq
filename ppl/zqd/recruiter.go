package zqd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brimsec/zq/api"
	"github.com/brimsec/zq/ppl/zqd/recruiter"
	"github.com/brimsec/zq/zqe"
	"go.uber.org/zap"
)

// handleRecruit and handleRegister interact with each other:
// completing a request to handleRecruit will unblock multiple
// open requests (long polls) to handleRegister.
// The mechanism for this is the "Callback" function in WorkerDetail.
// The callback is a closure in handleRecruit which will write
// to a channel which unblocks the request. This "recruited"
// channel is only used within the body of the handleRegister function.
func handleRecruit(c *Core, w http.ResponseWriter, r *http.Request) {
	var req api.RecruitRequest
	if !request(c, w, r, &req) {
		return
	}
	ws, err := c.workerPool.Recruit(req.NumberRequested)
	if err != nil {
		respondError(c, w, r, zqe.ErrInvalid(err))
		return
	}
	var workers []api.Worker
	for _, e := range ws {
		if e.Callback(recruiter.RecruitmentDetail{LoggingLabel: req.Label, NumberRequested: req.NumberRequested}) {
			workers = append(workers, api.Worker{Addr: e.Addr, NodeName: e.NodeName})
		}
	}
	respond(c, w, r, http.StatusOK, api.RecruitResponse{
		Workers: workers,
	})
}

func handleRegister(c *Core, w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest
	if !request(c, w, r, &req) {
		return
	}
	if req.Timeout <= 0 {
		respondError(c, w, r, zqe.E(zqe.Invalid, "required parameter timeout"))
		return
	}
	timer := time.NewTimer(time.Duration(req.Timeout) * time.Millisecond)
	recruited := make(chan recruiter.RecruitmentDetail)
	defer timer.Stop()
	cb := func(rd recruiter.RecruitmentDetail) bool {
		timer.Stop()
		select {
		case recruited <- rd:
		default:
			c.logger.Warn("Receiver not ready for recruited", zap.String("label", rd.LoggingLabel))
			return false
			// Logs on a race between /recruiter/recruit and req.Timeout.
			// If the receiver is not ready it means the worker has Deregistered.
			// Return false so worker is omitted from response.
		}
		return true
	}
	if err := c.workerPool.Register(req.Addr, req.NodeName, cb); err != nil {
		respondError(c, w, r, zqe.ErrInvalid(err))
		return
	}
	// directive is one of:
	//  "reserved"   indicates to the worker that is has been reserved by a root process.
	//  "reregister" indicates the request timed out without the worker being reserved,
	//               and the worker should send another register request.
	var directive string
	var isCanceled bool
	ctx := r.Context()
	select {
	case rd := <-recruited:
		c.requestLogger(r).Info("Worker recruited",
			zap.String("addr", req.Addr),
			zap.String("label", rd.LoggingLabel),
			zap.Int("count", rd.NumberRequested))
		directive = "reserved"
	case <-timer.C:
		c.requestLogger(r).Info("Worker should reregister", zap.String("addr", req.Addr))
		c.workerPool.Deregister(req.Addr)
		directive = "reregister"
	case <-ctx.Done():
		c.requestLogger(r).Info("HandleRegister context cancel")
		c.workerPool.Deregister(req.Addr)
		isCanceled = true
	}
	if !isCanceled {
		respond(c, w, r, http.StatusOK, api.RegisterResponse{Directive: directive})
	}
}

func handleRecruiterStats(c *Core, w http.ResponseWriter, r *http.Request) {
	respond(c, w, r, http.StatusOK, api.RecruiterStatsResponse{
		LenFreePool: c.workerPool.LenFreePool(),
		LenNodePool: c.workerPool.LenNodePool(),
	})
}

// handleListFree pretty prints the output because it is for manual trouble-shooting.
func handleListFree(c *Core, w http.ResponseWriter, r *http.Request) {
	ws := c.workerPool.ListFreePool()
	workers := make([]api.Worker, len(ws))
	for i, e := range ws {
		workers[i] = api.Worker{Addr: e.Addr, NodeName: e.NodeName}
	}
	body := api.RecruitResponse{
		Workers: workers,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(body); err != nil {
		c.requestLogger(r).Warn("Error writing response", zap.Error(err))
	}
}
