package coerce

import (
	"math"

	zed "github.com/brimdata/super"
)

func IntBits(typ zed.Type) int {
	switch typ.ID() {
	case zed.IDInt8:
		return 8
	case zed.IDInt16:
		return 16
	case zed.IDInt32:
		return 32
	case zed.IDInt64, zed.IDDuration, zed.IDTime:
		return 64
	default:
		panic(typ)
	}
}

func UintBits(typ zed.Type) int {
	switch typ.ID() {
	case zed.IDUint8:
		return 8
	case zed.IDUint16:
		return 16
	case zed.IDUint32:
		return 32
	case zed.IDUint64:
		return 64
	default:
		panic(typ)
	}
}

func FromIntOverflowCheck(from, to zed.Type) (int64, int64, bool) {
	fromid := from.ID()
	if !zed.IsSigned(fromid) {
		panic("from: not int")
	}
	switch toid := to.ID(); {
	case zed.IsSigned(toid):
		check := fromid > toid
		switch toid {
		case zed.IDInt8:
			return math.MinInt8, math.MaxInt8, check
		case zed.IDInt16:
			return math.MinInt16, math.MaxInt16, check
		case zed.IDInt32:
			return math.MinInt32, math.MaxInt32, check
		default:
			return 0, 0, false
		}
	case zed.IsUnsigned(toid):
		// You have to check no matter what b/c of negative
		switch toid {
		case zed.IDUint8:
			return 0, math.MaxUint8, true
		case zed.IDUint16:
			return 0, math.MaxUint16, true
		case zed.IDUint32:
			return 0, math.MaxUint32, true
		default:
			return 0, math.MaxInt64, true
		}
	case zed.IsFloat(toid):
		return 0, 0, false
	default:
		panic("to: non-numeric")
	}
}

func FromUintOverflowCheck(from, to zed.Type) (uint64, bool) {
	fromid := from.ID()
	if !zed.IsUnsigned(fromid) {
		panic("from: not uint")
	}
	switch toid := to.ID(); {
	case zed.IsUnsigned(toid):
		check := fromid > toid
		switch toid {
		case zed.IDUint8:
			return math.MaxUint8, check
		case zed.IDUint16:
			return math.MaxUint16, check
		case zed.IDUint32:
			return math.MaxUint32, check
		default:
			return 0, false
		}
	case zed.IsSigned(toid):
		switch toid {
		case zed.IDInt8:
			return math.MaxInt8, true
		case zed.IDInt16:
			return math.MaxInt16, fromid >= zed.IDUint16
		case zed.IDInt32:
			return math.MaxInt32, fromid >= zed.IDUint32
		default:
			return math.MaxInt64, fromid == zed.IDUint64
		}
	case zed.IsFloat(toid):
		// float has to be checked because < 0
		return math.MaxUint64, true
	default:
		panic("to: non-numeric")
	}
}

func FromFloatOverflowCheck(from, to zed.Type) (float64, float64, bool) {
	fromid := from.ID()
	if !zed.IsFloat(fromid) {
		panic("from: not float")
	}
	toid := to.ID()
	switch toid {
	case zed.IDUint8:
		return 0, math.MaxUint8, true
	case zed.IDUint16:
		return 0, math.MaxUint16, true
	case zed.IDUint32:
		return 0, math.MaxUint32, true
	case zed.IDUint64:
		return 0, math.MaxUint64, true
	case zed.IDInt8:
		return math.MinInt8, math.MaxInt8, true
	case zed.IDInt16:
		return math.MinInt16, math.MaxInt16, true
	case zed.IDInt32:
		return math.MinInt32, math.MaxInt32, true
	case zed.IDInt64, zed.IDTime, zed.IDDuration:
		return math.MinInt64, math.MaxInt64, true
	case zed.IDFloat16, zed.IDFloat32, zed.IDFloat64:
		return 0, 0, false
	default:
		panic("to: non-numeric")
	}
}
