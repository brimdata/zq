package zjsonio

import (
	"errors"
	"fmt"

	"github.com/brimdata/zed/compiler/ast"
	"github.com/brimdata/zed/zng"
	"github.com/brimdata/zed/zson"
)

type encoder map[zng.Type]string

func (e encoder) encodeType(zctx *zson.Context, typ zng.Type) ast.Type {
	if name, ok := e[typ]; ok {
		return &ast.TypeName{
			Kind: "typename",
			Name: name,
		}
	}
	switch typ := typ.(type) {
	case *zng.TypeAlias:
		name := typ.Name
		t := e.encodeType(zctx, typ.Type)
		e[typ] = name
		return &ast.TypeDef{
			Kind: "typedef",
			Name: name,
			Type: t,
		}
	case *zng.TypeRecord:
		return e.encodeTypeRecord(zctx, typ)
	case *zng.TypeArray:
		return &ast.TypeArray{
			Kind: "array",
			Type: e.encodeType(zctx, typ.Type),
		}
	case *zng.TypeSet:
		return &ast.TypeSet{
			Kind: "set",
			Type: e.encodeType(zctx, typ.Type),
		}
	case *zng.TypeUnion:
		return e.encodeTypeUnion(zctx, typ)
	case *zng.TypeEnum:
		return e.encodeTypeEnum(zctx, typ)
	case *zng.TypeMap:
		return &ast.TypeMap{
			Kind:    "map",
			KeyType: e.encodeType(zctx, typ.KeyType),
			ValType: e.encodeType(zctx, typ.ValType),
		}
	default:
		return &ast.TypePrimitive{
			Kind: "primitive",
			Name: typ.ZSON(),
		}
	}
}

func (e encoder) encodeTypeRecord(zctx *zson.Context, typ *zng.TypeRecord) *ast.TypeRecord {
	var fields []ast.TypeField
	for _, c := range typ.Columns {
		typ := e.encodeType(zctx, c.Type)
		fields = append(fields, ast.TypeField{c.Name, typ})
	}
	return &ast.TypeRecord{
		Kind:   "record",
		Fields: fields,
	}
}

func (e encoder) encodeTypeEnum(zctx *zson.Context, typ *zng.TypeEnum) *ast.TypeEnum {
	panic("issue 2508")
}

func (e encoder) encodeTypeUnion(zctx *zson.Context, union *zng.TypeUnion) *ast.TypeUnion {
	var types []ast.Type
	for _, t := range union.Types {
		types = append(types, e.encodeType(zctx, t))
	}
	return &ast.TypeUnion{
		Kind:  "union",
		Types: types,
	}
}

type decoder map[string]zng.Type

func (d decoder) decodeType(zctx *zson.Context, typ ast.Type) (zng.Type, error) {
	switch typ := typ.(type) {
	case *ast.TypeRecord:
		return d.decodeTypeRecord(zctx, typ)
	case *ast.TypeArray:
		t, err := d.decodeType(zctx, typ.Type)
		if err != nil {
			return nil, err
		}
		return zctx.LookupTypeArray(t), nil
	case *ast.TypeSet:
		t, err := d.decodeType(zctx, typ.Type)
		if err != nil {
			return nil, err
		}
		return zctx.LookupTypeSet(t), nil
	case *ast.TypeUnion:
		return d.decodeTypeUnion(zctx, typ)
	case *ast.TypeEnum:
		return d.decodeTypeEnum(zctx, typ)
	case *ast.TypeMap:
		return d.decodeTypeMap(zctx, typ)
	case *ast.TypeName:
		t := zctx.LookupTypeDef(typ.Name)
		if typ == nil {
			return nil, fmt.Errorf("ZJSON decoder: no such type name: %s", typ.Name)
		}
		return t, nil
	case *ast.TypeDef:
		t, err := d.decodeType(zctx, typ.Type)
		if err != nil {
			return nil, err
		}
		d[typ.Name] = t
		if !zng.IsIdentifier(typ.Name) {
			return t, nil
		}
		return zctx.LookupTypeAlias(typ.Name, t)
	case *ast.TypePrimitive:
		t, err := zctx.LookupByName(typ.Name)
		if err != nil {
			return nil, errors.New("ZJSON unknown type: " + typ.Name)
		}
		return t, nil
	}
	return nil, fmt.Errorf("ZJSON unknown type: %T", typ)
}

func (d decoder) decodeTypeRecord(zctx *zson.Context, typ *ast.TypeRecord) (*zng.TypeRecord, error) {
	columns := make([]zng.Column, 0, len(typ.Fields))
	for _, field := range typ.Fields {
		typ, err := d.decodeType(zctx, field.Type)
		if err != nil {
			return nil, err
		}
		column := zng.Column{
			Name: field.Name,
			Type: typ,
		}
		columns = append(columns, column)
	}
	return zctx.LookupTypeRecord(columns)
}

func (d decoder) decodeTypeUnion(zctx *zson.Context, union *ast.TypeUnion) (*zng.TypeUnion, error) {
	var types []zng.Type
	for _, t := range union.Types {
		typ, err := d.decodeType(zctx, t)
		if err != nil {
			return nil, err
		}
		types = append(types, typ)
	}
	return zctx.LookupTypeUnion(types), nil
}

func (d decoder) decodeTypeMap(zctx *zson.Context, m *ast.TypeMap) (*zng.TypeMap, error) {
	keyType, err := d.decodeType(zctx, m.KeyType)
	if err != nil {
		return nil, err
	}
	valType, err := d.decodeType(zctx, m.ValType)
	if err != nil {
		return nil, err
	}
	return zctx.LookupTypeMap(keyType, valType), nil
}

func (d decoder) decodeTypeEnum(zctx *zson.Context, enum *ast.TypeEnum) (*zng.TypeEnum, error) {
	return nil, errors.New("TBD: issue #2508")
}
