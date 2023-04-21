package meta

import (
	"context"
	"errors"
	"io"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/lake"
	"github.com/brimdata/zed/lake/data"
	"github.com/brimdata/zed/runtime/expr"
	"github.com/brimdata/zed/runtime/op"
	"github.com/brimdata/zed/runtime/op/merge"
	"github.com/brimdata/zed/zbuf"
	"github.com/brimdata/zed/zio/zngio"
	"github.com/brimdata/zed/zson"
)

// SequenceScanner implements an op that pulls metadata partitions to scan
// from its parent and for each partition, scans the object.
type SequenceScanner struct {
	parent      zbuf.Puller
	current     zbuf.Puller
	filter      zbuf.Filter
	pruner      expr.Evaluator
	octx        *op.Context
	pool        *lake.Pool
	progress    *zbuf.Progress
	unmarshaler *zson.UnmarshalZNGContext
	done        bool
	err         error
}

func NewSequenceScanner(octx *op.Context, parent zbuf.Puller, pool *lake.Pool, filter zbuf.Filter, pruner expr.Evaluator, progress *zbuf.Progress) *SequenceScanner {
	return &SequenceScanner{
		octx:        octx,
		parent:      parent,
		filter:      filter,
		pruner:      pruner,
		pool:        pool,
		progress:    progress,
		unmarshaler: zson.NewZNGUnmarshaler(),
	}
}

func (s *SequenceScanner) Pull(done bool) (zbuf.Batch, error) {
	if s.done {
		return nil, s.err
	}
	if done {
		if s.current != nil {
			_, err := s.current.Pull(true)
			s.close(err)
			s.current = nil
		}
		return nil, s.err
	}
	for {
		if s.current == nil {
			if s.parent == nil { //XXX
				s.close(nil)
				return nil, nil
			}
			// Pull the next partition from the parent snapshot and
			// set up the next scanner to pull from.
			var part Partition
			ok, err := nextPartition(s.parent, &part, s.unmarshaler)
			if !ok || err != nil {
				s.close(err)
				return nil, err
			}
			s.current, err = newSortedPartitionScanner(s, part)
			if err != nil {
				s.close(err)
				return nil, err
			}
		}
		batch, err := s.current.Pull(false)
		if err != nil {
			s.close(err)
			return nil, err
		}
		if batch != nil {
			return batch, nil
		}
		s.current = nil
	}
}

func (s *SequenceScanner) close(err error) {
	s.err = err
	s.done = true
}

func nextPartition(puller zbuf.Puller, part *Partition, u *zson.UnmarshalZNGContext) (bool, error) {
	batch, err := puller.Pull(false)
	if batch == nil || err != nil {
		return false, err
	}
	vals := batch.Values()
	if len(vals) != 1 {
		// We currently support only one partition per batch.
		return false, errors.New("system error: SequenceScanner encountered multi-valued batch")
	}
	return true, u.Unmarshal(&vals[0], part)
}

func newSortedPartitionScanner(p *SequenceScanner, part Partition) (zbuf.Puller, error) {
	pullers := make([]zbuf.Puller, 0, len(part.Objects))
	pullersDone := func() {
		for _, puller := range pullers {
			puller.Pull(true)
		}
	}
	for _, object := range part.Objects {
		s, err := newObjectScanner(p.octx.Context, p.octx.Zctx, p.pool, object, p.filter, p.pruner, p.progress)
		if err != nil {
			pullersDone()
			return nil, err
		}
		pullers = append(pullers, s)
	}
	if len(pullers) == 1 {
		return pullers[0], nil
	}
	return merge.New(p.octx.Context, pullers, lake.ImportComparator(p.octx.Zctx, p.pool).Compare), nil
}

func newObjectScanner(ctx context.Context, zctx *zed.Context, pool *lake.Pool, object *data.Object, filter zbuf.Filter, pruner expr.Evaluator, progress *zbuf.Progress) (zbuf.Puller, error) {
	ranges, err := data.LookupSeekRange(ctx, pool.Storage(), pool.DataPath, object, pruner)
	if err != nil {
		return nil, err
	}
	rc, err := object.NewReader(ctx, pool.Storage(), pool.DataPath, ranges)
	if err != nil {
		return nil, err
	}
	scanner, err := zngio.NewReader(zctx, rc).NewScanner(ctx, filter)
	if err != nil {
		rc.Close()
		return nil, err
	}
	return &statScanner{
		scanner:  scanner,
		closer:   rc,
		progress: progress,
	}, nil
}

type statScanner struct {
	scanner  zbuf.Scanner
	closer   io.Closer
	err      error
	progress *zbuf.Progress
}

func (s *statScanner) Pull(done bool) (zbuf.Batch, error) {
	if s.scanner == nil {
		return nil, s.err
	}
	batch, err := s.scanner.Pull(done)
	if batch == nil || err != nil {
		s.progress.Add(s.scanner.Progress())
		if err2 := s.closer.Close(); err == nil {
			err = err2
		}
		s.err = err
		s.scanner = nil
	}
	return batch, err
}
