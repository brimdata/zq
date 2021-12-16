package switcher

import (
	"sync"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/expr"
	"github.com/brimdata/zed/proc"
	"github.com/brimdata/zed/zbuf"
)

type request struct {
	i  int
	ch chan<- proc.Result
}

type Switcher struct {
	parent     proc.Interface
	once       sync.Once
	n          int
	filters    []expr.Filter
	requestsCh chan request
	requests   []chan<- proc.Result
}

func New(parent proc.Interface) *Switcher {
	return &Switcher{
		parent:     parent,
		requestsCh: make(chan request),
	}
}

func (s *Switcher) Add(filt expr.Filter) (chan request, int) {
	s.filters = append(s.filters, filt)
	s.requests = append(s.requests, nil)
	id := s.n
	s.n++
	return s.requestsCh, id
}

// gather ensures that we have a request for each active downstream.
func (s *Switcher) gather() {
	var nready int
	for _, r := range s.requests {
		if r != nil {
			nready++
		}
	}
	for nready < s.n {
		req := <-s.requestsCh
		if req.ch == nil {
			s.n--
			continue
		}
		nready++
		if s.requests[req.i] != nil {
			panic("Switcher bug")
		}
		s.requests[req.i] = req.ch
	}
}

func (s *Switcher) run() {
	records := make([][]zed.Value, s.n)
	results := make([]proc.Result, s.n)
	for {
		s.gather()
		if s.n == 0 {
			break
		}
		batch, err := s.parent.Pull()
		if proc.EOS(batch, err) {
			s.sendEOS(proc.Result{batch, err})
			continue
		}
		ctx := batch.Context()
		vals := batch.Values()
		for i := range vals {
			if j := s.match(ctx, &vals[i]); j >= 0 {
				if records[j] == nil {
					records[j] = make([]zed.Value, 0, len(vals))
				}
				records[j] = append(records[j], vals[i])
			}
		}
		for i := range records {
			if records[i] != nil {
				results[i] = proc.Result{Batch: zbuf.NewArray(records[i])}
				records[i] = nil
			}
		}
		s.send(results)
		batch.Unref()
	}
	s.parent.Done()
}

func (s *Switcher) match(ctx expr.Context, rec *zed.Value) int {
	for i, f := range s.filters {
		if f(ctx, rec) {
			return i
		}
	}
	return -1
}

func (s *Switcher) sendEOS(result proc.Result) {
	for i := range s.requests {
		if s.requests[i] != nil {
			s.requests[i] <- result
			s.requests[i] = nil
		}
	}
}

func (s *Switcher) send(results []proc.Result) {
	for i, ch := range s.requests {
		if results[i].Batch != nil {
			results[i].Batch.Ref()
			ch <- results[i]
			s.requests[i] = nil
			results[i].Batch = nil
		}
	}
}

type Proc struct {
	id     int
	reqCh  chan request
	ch     chan proc.Result
	parent *Switcher
}

func (s *Switcher) NewProc(filt expr.Filter) *Proc {
	p := &Proc{
		ch:     make(chan proc.Result),
		parent: s,
	}
	p.reqCh, p.id = s.Add(filt)
	return p
}

func (p *Proc) Pull() (zbuf.Batch, error) {
	p.parent.once.Do(func() {
		go p.parent.run()
	})
	if p.ch == nil {
		return nil, nil
	}
	// Send parent a request, then read the result.
	p.reqCh <- request{p.id, p.ch}
	result := <-p.ch
	if proc.EOS(result.Batch, result.Err) {
		p.Done()
	}
	return result.Batch, result.Err
}

func (p *Proc) Done() {
	// Signal to our parent Switcher that this path is done by
	// sending a request with a nil channel object.  We go ahead
	// and mark the Proc done by setting its channel to nil in
	// case a spurious Pull() is called, but this should not
	// happen.
	if p.ch != nil {
		p.ch = nil
		p.reqCh <- request{p.id, p.ch}
	}
}
