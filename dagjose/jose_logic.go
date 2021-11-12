package dagjose

import (
	"encoding/base64"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/ipld/go-ipld-prime/schema"
)

func (n Base64String) String() string {
	return toBase64Url(n.x)
}
func (_Base64String__Prototype) fromString(w *_Base64String, v string) error {
	*w = _Base64String{toBase64Url(v)}
	return nil
}
func (_Base64String__Prototype) FromString(v string) (Base64String, error) {
	n := _Base64String{toBase64Url(v)}
	return &n, nil
}

type _Base64String__Maybe struct {
	m schema.Maybe
	v _Base64String
}
type MaybeBase64String = *_Base64String__Maybe

func (m MaybeBase64String) IsNull() bool {
	return m.m == schema.Maybe_Null
}
func (m MaybeBase64String) IsAbsent() bool {
	return m.m == schema.Maybe_Absent
}
func (m MaybeBase64String) Exists() bool {
	return m.m == schema.Maybe_Value
}
func (m MaybeBase64String) AsNode() datamodel.Node {
	switch m.m {
	case schema.Maybe_Absent:
		return datamodel.Absent
	case schema.Maybe_Null:
		return datamodel.Null
	case schema.Maybe_Value:
		return &m.v
	default:
		panic("unreachable")
	}
}
func (m MaybeBase64String) Must() Base64String {
	if !m.Exists() {
		panic("unbox of a maybe rejected")
	}
	return &m.v
}

var _ datamodel.Node = (Base64String)(&_Base64String{})
var _ schema.TypedNode = (Base64String)(&_Base64String{})

func (Base64String) Kind() datamodel.Kind {
	return datamodel.Kind_String
}
func (Base64String) LookupByString(string) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.LookupByString("")
}
func (Base64String) LookupByNode(datamodel.Node) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.LookupByNode(nil)
}
func (Base64String) LookupByIndex(idx int64) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.LookupByIndex(0)
}
func (Base64String) LookupBySegment(seg datamodel.PathSegment) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.LookupBySegment(seg)
}
func (Base64String) MapIterator() datamodel.MapIterator {
	return nil
}
func (Base64String) ListIterator() datamodel.ListIterator {
	return nil
}
func (Base64String) Length() int64 {
	return -1
}
func (Base64String) IsAbsent() bool {
	return false
}
func (Base64String) IsNull() bool {
	return false
}
func (Base64String) AsBool() (bool, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.AsBool()
}
func (Base64String) AsInt() (int64, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.AsInt()
}
func (Base64String) AsFloat() (float64, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.AsFloat()
}
func (n Base64String) AsString() (string, error) {
	return toBase64Url(n.x), nil
}
func (Base64String) AsBytes() ([]byte, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.AsBytes()
}
func (Base64String) AsLink() (datamodel.Link, error) {
	return mixins.String{TypeName: "dagjose.Base64String"}.AsLink()
}
func (Base64String) Prototype() datamodel.NodePrototype {
	return _Base64String__Prototype{}
}

type _Base64String__Prototype struct{}

func (_Base64String__Prototype) NewBuilder() datamodel.NodeBuilder {
	var nb _Base64String__Builder
	nb.Reset()
	return &nb
}

type _Base64String__Builder struct {
	_Base64String__Assembler
}

func (nb *_Base64String__Builder) Build() datamodel.Node {
	if *nb.m != schema.Maybe_Value {
		panic("invalid state: cannot call Build on an assembler that's not finished")
	}
	return nb.w
}
func (nb *_Base64String__Builder) Reset() {
	var w _Base64String
	var m schema.Maybe
	*nb = _Base64String__Builder{_Base64String__Assembler{w: &w, m: &m}}
}

type _Base64String__Assembler struct {
	w *_Base64String
	m *schema.Maybe
}

func (na *_Base64String__Assembler) reset() {}
func (_Base64String__Assembler) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.BeginMap(0)
}
func (_Base64String__Assembler) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.BeginList(0)
}
func (na *_Base64String__Assembler) AssignNull() error {
	switch *na.m {
	case allowNull:
		*na.m = schema.Maybe_Null
		return nil
	case schema.Maybe_Absent:
		return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.AssignNull()
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	panic("unreachable")
}
func (_Base64String__Assembler) AssignBool(bool) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.AssignBool(false)
}
func (_Base64String__Assembler) AssignInt(int64) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.AssignInt(0)
}
func (_Base64String__Assembler) AssignFloat(float64) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.AssignFloat(0)
}
func (na *_Base64String__Assembler) AssignString(v string) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	decoded, err := fromBase64Url(v)
	if err != nil {
		return err
	}
	na.w.x = decoded
	*na.m = schema.Maybe_Value
	return nil
}
func (na *_Base64String__Assembler) AssignBytes(v []byte) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	na.w.x = string(v)
	*na.m = schema.Maybe_Value
	return nil
}
func (_Base64String__Assembler) AssignLink(datamodel.Link) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64String"}.AssignLink(nil)
}
func (na *_Base64String__Assembler) AssignNode(v datamodel.Node) error {
	if v.IsNull() {
		return na.AssignNull()
	}
	if v2, ok := v.(*_Base64String); ok {
		switch *na.m {
		case schema.Maybe_Value, schema.Maybe_Null:
			panic("invalid state: cannot assign into assembler that's already finished")
		}
		*na.w = *v2
		*na.m = schema.Maybe_Value
		return nil
	}
	if v2, err := v.AsString(); err != nil {
		return err
	} else {
		return na.AssignString(v2)
	}
}
func (_Base64String__Assembler) Prototype() datamodel.NodePrototype {
	return _Base64String__Prototype{}
}
func (Base64String) Type() schema.Type {
	return nil /*TODO:typelit*/
}
func (n Base64String) Representation() datamodel.Node {
	return (*_Base64String__Repr)(n)
}

type _Base64String__Repr = _Base64String

var _ datamodel.Node = &_Base64String__Repr{}

type _Base64String__ReprPrototype = _Base64String__Prototype
type _Base64String__ReprAssembler = _Base64String__Assembler

func toBase64Url(decoded string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(decoded))
}

func fromBase64Url(encoded string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	return string(decoded), err
}
