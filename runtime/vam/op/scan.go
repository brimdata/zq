package op

import (
	"errors"
	"fmt"
	"sync"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/lake"
	"github.com/brimdata/zed/lake/data"
	"github.com/brimdata/zed/pkg/field"
	"github.com/brimdata/zed/runtime"
	"github.com/brimdata/zed/runtime/sam/expr"
	"github.com/brimdata/zed/runtime/vcache"
	"github.com/brimdata/zed/vector"
	"github.com/brimdata/zed/zbuf"
	"github.com/brimdata/zed/zson"
)

type Scanner struct {
	parent      zbuf.Puller
	pruner      expr.Evaluator
	rctx        *runtime.Context
	pool        *lake.Pool
	once        sync.Once
	projection  vcache.Path
	cache       *vcache.Cache
	progress    *zbuf.Progress
	unmarshaler *zson.UnmarshalZNGContext
	resultCh    chan result
	doneCh      chan struct{}
}

var _ vector.Puller = (*Scanner)(nil)

func NewScanner(rctx *runtime.Context, cache *vcache.Cache, parent zbuf.Puller, pool *lake.Pool, paths []field.Path, pruner expr.Evaluator, progress *zbuf.Progress) *Scanner {
	return &Scanner{
		cache:       cache,
		rctx:        rctx,
		parent:      parent,
		pruner:      pruner,
		pool:        pool,
		projection:  vcache.NewProjection(paths),
		progress:    progress,
		unmarshaler: zson.NewZNGUnmarshaler(),
		doneCh:      make(chan struct{}),
		resultCh:    make(chan result),
	}
}

// XXX we need vector scannerstats and means to update them here.

// XXX change this to pull/load vector by each type within an object and
// return an object containing the overall projection, which might be a record
// or could just be a single vector.  the downstream operator has to be
// configured to expect it, e.g., project x:=a.b,y:=a.b.c (like cut but in vspace)
// this would be Record{x:(proj a.b),y:(proj:a.b.c)} so the elements would be
// single fields.  For each object/type that matches the projection we would make
// a Record vec and let GC reclaim them.  Note if a col is missing, it's a constant
// vector of error("missing").

func (s *Scanner) Pull(done bool) (vector.Any, error) {
	s.once.Do(func() {
		// Block p.ctx's cancel function until p.run finishes its
		// cleanup.
		s.rctx.WaitGroup.Add(1)
		go s.run()
	})
	if done {
		select {
		case s.doneCh <- struct{}{}:
			return nil, nil
		case <-s.rctx.Done():
			return nil, s.rctx.Err()
		}
	}
	if r, ok := <-s.resultCh; ok {
		return r.vector, r.err
	}
	return nil, s.rctx.Err()
}

func (s *Scanner) run() {
	defer func() {
		s.rctx.WaitGroup.Done()
	}()
	for {
		//XXX should make an object puller that wraps this...
		batch, err := s.parent.Pull(false)
		if batch == nil || err != nil {
			s.sendResult(nil, err)
			return
		}
		vals := batch.Values()
		if len(vals) != 1 {
			// We require exactly one data object per pull.
			err := errors.New("system error: vam.Scanner encountered multi-valued batch")
			s.sendResult(nil, err)
			return
		}
		named, ok := vals[0].Type().(*zed.TypeNamed)
		if !ok {
			s.sendResult(nil, fmt.Errorf("system error: vam.Scanner encountered unnamed object: %s", zson.String(vals[0])))
			return
		}
		if named.Name != "data.Object" {
			s.sendResult(nil, fmt.Errorf("system error: vam.Scanner encountered unnamed object: %q", named.Name))
			return
		}
		var meta data.Object
		if err := s.unmarshaler.Unmarshal(vals[0], &meta); err != nil {
			s.sendResult(nil, fmt.Errorf("system error: vam.Scanner could not unmarshal value: %q", zson.String(vals[0])))
			return
		}
		object, err := s.cache.Fetch(s.rctx.Context, meta.VectorURI(s.pool.DataPath), meta.ID)
		if err != nil {
			s.sendResult(nil, err)
			return
		}
		vec, err := object.Fetch(s.rctx.Zctx, s.projection)
		s.sendResult(vec, err)
		if err != nil {
			return
		}
	}
}

func (s *Scanner) sendResult(vec vector.Any, err error) (bool, bool) {
	select {
	case s.resultCh <- result{vec, err}:
		return false, true
	case <-s.doneCh:
		b, pullErr := s.parent.Pull(true)
		if err == nil {
			err = pullErr
		}
		if err != nil {
			select {
			case s.resultCh <- result{err: err}:
				return true, false
			case <-s.rctx.Done():
				return false, false
			}
		}
		if b != nil {
			b.Unref()
		}
		return true, true
	case <-s.rctx.Done():
		return false, false
	}
}

type result struct {
	vector vector.Any
	err    error //XXX go err vs vector.Any err?
}