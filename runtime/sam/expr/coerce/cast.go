package coerce

import (
	"strconv"

	zed "github.com/brimdata/super"
	"github.com/brimdata/super/pkg/byteconv"
	"github.com/x448/float16"
)

func ToUint(val zed.Value, typUint zed.Type) (uint64, bool) {
	val = val.Under()
	switch id := val.Type().ID(); {
	case zed.IsUnsigned(id):
		v := val.Uint()
		max, check := FromUintOverflowCheck(val.Type(), typUint)
		return v, !check || v <= max
	case zed.IsSigned(id):
		v := val.Int()
		min, max, check := FromIntOverflowCheck(val.Type(), typUint)
		return uint64(v), !check || (v >= min && v <= max)
	case zed.IsFloat(id):
		v := val.Float()
		min, max, check := FromFloatOverflowCheck(val.Type(), typUint)
		return uint64(v), !check || v >= min && v <= max
	case id == zed.IDString:
		v, err := strconv.ParseUint(val.AsString(), 10, UintBits(typUint))
		return v, err == nil
	}
	return 0, false
}

func ToInt(val zed.Value, typInt zed.Type) (int64, bool) {
	val = val.Under()
	switch id := val.Type().ID(); {
	case zed.IsUnsigned(id):
		v := val.Uint()
		max, check := FromUintOverflowCheck(val.Type(), typInt)
		return int64(v), !check || v <= max
	case zed.IsSigned(id):
		v := val.Int()
		min, max, check := FromIntOverflowCheck(val.Type(), typInt)
		return v, !check || v >= min && v <= max
	case zed.IsFloat(id):
		v := val.Float()
		min, max, check := FromFloatOverflowCheck(val.Type(), typInt)
		return int64(v), !check || v >= min && v <= max
	case id == zed.IDString:
		v, err := strconv.ParseInt(val.AsString(), 10, IntBits(typInt))
		return v, err == nil
	}
	return 0, false
}

func ToFloat(val zed.Value, typ zed.Type) (float64, bool) {
	var v float64
	val = val.Under()
	fromId := val.Type().ID()
	switch {
	case zed.IsUnsigned(fromId):
		v = float64(val.Uint())
	case zed.IsSigned(fromId):
		v = float64(val.Int())
	case zed.IsFloat(fromId):
		v = val.Float()
	case fromId == zed.IDString:
		var err error
		if v, err = byteconv.ParseFloat64(val.Bytes()); err != nil {
			return v, false
		}
	}
	switch typ.ID() {
	case zed.IDFloat16:
		if fromId != zed.IDFloat16 {
			f16 := float16.Fromfloat32(float32(v))
			return float64(f16.Float32()), true
		}
	case zed.IDFloat32:
		if fromId != zed.IDFloat16 && fromId != zed.IDFloat32 {
			return float64(float32(v)), true
		}
	}
	return v, true
}

func ToBool(val zed.Value) (bool, bool) {
	val = val.Under()
	if val.IsString() {
		v, err := byteconv.ParseBool(val.Bytes())
		return v, err == nil
	}
	v, ok := ToInt(val, zed.TypeInt64)
	return v != 0, ok
}
