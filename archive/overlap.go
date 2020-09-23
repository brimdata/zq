package archive

import (
	"sort"

	"github.com/brimsec/zq/pkg/nano"
	"github.com/brimsec/zq/zbuf"
)

// mergeChunksToSpans takes a set of Chunks with possibly overlapping spans,
// and returns an ordered list of SpanInfos, whose spans will be bounded by
// filter, and where each spanInfo contains one or more Chunks whose data
// falls into the spanInfo's span.
func mergeChunksToSpans(chunks []Chunk, dir zbuf.Direction, filter nano.Span) []spanInfo {
	var siChunks []Chunk // accumulating chunks for next spanInfo
	var siFirst nano.Ts  // first timestamp for next spanInfo
	var result []spanInfo
	boundaries(chunks, dir, func(ts nano.Ts, firstChunks, lastChunks []Chunk) {
		if len(firstChunks) > 0 {
			// ts is the 'First' timestamp for these chunks.
			if len(siChunks) > 0 {
				// We have accumulated chunks; create a span with them whose
				// last timestamp was just before ts.
				siSpan := closedSpan(siFirst, prevTs(ts, dir))
				if filter.Overlaps(siSpan) {
					result = append(result, spanInfo{
						span:   filter.Intersect(siSpan),
						chunks: copyChunks(siChunks, nil),
					})
				}
			}
			// Accumulate these chunks whose first timestamp is ts.
			siChunks = append(siChunks, firstChunks...)
			siFirst = ts
		}
		if len(lastChunks) > 0 {
			// ts is the 'Last' timestamp for these chunks.
			siSpan := closedSpan(siFirst, ts)
			if filter.Overlaps(siSpan) {
				result = append(result, spanInfo{
					span:   filter.Intersect(siSpan),
					chunks: copyChunks(siChunks, nil),
				})
			}
			// Drop the chunks that ended from our accumulation.
			siChunks = copyChunks(siChunks, lastChunks)
			siFirst = nextTs(ts, dir)
		}
	})
	return result
}

func copyChunks(src []Chunk, skip []Chunk) (dst []Chunk) {
outer:
	for i := range src {
		for j := range skip {
			if src[i].Id == skip[j].Id {
				continue outer
			}
		}
		dst = append(dst, src[i])
	}
	return
}

// closedSpan returns a span for the closed interval of [x,y].
func closedSpan(x, y nano.Ts) nano.Span {
	return nano.Span{Ts: x, Dur: 1}.Union(nano.Span{Ts: y, Dur: 1})
}

// spanToFirstLast returns the timestamps that whose closed interval
// is represented by span s. It assumes s.Dur is greater than zero.
func spanToFirstLast(dir zbuf.Direction, s nano.Span) (first, last nano.Ts) {
	a := s.Ts
	b := s.End() - 1
	if dir == zbuf.DirTimeForward {
		return a, b
	}
	return b, a
}

func nextTs(ts nano.Ts, dir zbuf.Direction) nano.Ts {
	if dir == zbuf.DirTimeForward {
		return ts + 1
	}
	return ts - 1
}

func prevTs(ts nano.Ts, dir zbuf.Direction) nano.Ts {
	if dir == zbuf.DirTimeForward {
		return ts - 1
	}
	return ts + 1
}

type point struct {
	idx   int
	first bool
	ts    nano.Ts
}

// boundaries sorts the given chunks, then calls fn with each timestamp that
// acts as a first and/or last timestamp of one or more of the chunks.
func boundaries(chunks []Chunk, dir zbuf.Direction, fn func(ts nano.Ts, firstChunks, lastChunks []Chunk)) {
	points := make([]point, 2*len(chunks))
	for i, c := range chunks {
		points[2*i] = point{idx: i, first: true, ts: c.First}
		points[2*i+1] = point{idx: i, ts: c.Last}
	}
	sort.Slice(points, func(i, j int) bool {
		return chunkTsCompare(dir, points[i].ts, chunks[points[i].idx].Id, points[j].ts, chunks[points[j].idx].Id)
	})
	firstChunks := make([]Chunk, 0, len(chunks))
	lastChunks := make([]Chunk, 0, len(chunks))
	for i := 0; i < len(points); {
		j := i + 1
		for ; j < len(points); j++ {
			if points[i].ts != points[j].ts {
				break
			}
		}
		firstChunks = firstChunks[:0]
		lastChunks = lastChunks[:0]
		for _, p := range points[i:j] {
			if p.first {
				firstChunks = append(firstChunks, chunks[p.idx])
			} else {
				lastChunks = append(lastChunks, chunks[p.idx])
			}
		}
		ts := points[i].ts
		i = j
		fn(ts, firstChunks, lastChunks)
	}
}
