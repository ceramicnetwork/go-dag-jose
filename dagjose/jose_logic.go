package dagjose

import (
	"encoding/base64"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/ipld/go-ipld-prime/schema"
)

// String2Bytes matches the IPLD Schema type "String2Bytes".  It has string kind.
type String2Bytes = *_String2Bytes

// TODO
type _String2Bytes struct{ x []byte }

func (n String2Bytes) String() string {
	return bytesToString(n.x)
}

func (_String2Bytes__Prototype) fromString(w *_String2Bytes, v string) error {
	rawBytes, err := stringToBytes(v)
	if err != nil {
		return err
	}
	*w = _String2Bytes{rawBytes}
	return nil
}

func (n String2Bytes) Bytes() []byte {
	return n.x
}

func (_String2Bytes__Prototype) fromBytes(w *_String2Bytes, v []byte) error {
	*w = _String2Bytes{v}
	return nil
}

func (_String2Bytes__Prototype) FromBytes(v []byte) (String2Bytes, error) {
	n := _String2Bytes{v}
	return &n, nil
}

type _String2Bytes__Maybe struct {
	m schema.Maybe
	v _String2Bytes
}

type MaybeString2Bytes = *_String2Bytes__Maybe

func (m MaybeString2Bytes) IsNull() bool {
	return m.m == schema.Maybe_Null
}

func (m MaybeString2Bytes) IsAbsent() bool {
	return m.m == schema.Maybe_Absent
}

func (m MaybeString2Bytes) Exists() bool {
	return m.m == schema.Maybe_Value
}

func (m MaybeString2Bytes) AsNode() datamodel.Node {
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

func (m MaybeString2Bytes) Must() String2Bytes {
	if !m.Exists() {
		panic("unbox of a maybe rejected")
	}
	return &m.v
}

var _ datamodel.Node = (String2Bytes)(&_String2Bytes{})
var _ schema.TypedNode = (String2Bytes)(&_String2Bytes{})

func (String2Bytes) Kind() datamodel.Kind {
	return datamodel.Kind_String
}

func (String2Bytes) LookupByString(string) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.LookupByString("")
}

func (String2Bytes) LookupByNode(datamodel.Node) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.LookupByNode(nil)
}

func (String2Bytes) LookupByIndex(idx int64) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.LookupByIndex(0)
}

func (String2Bytes) LookupBySegment(seg datamodel.PathSegment) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.LookupBySegment(seg)
}

func (String2Bytes) MapIterator() datamodel.MapIterator {
	return nil
}

func (String2Bytes) ListIterator() datamodel.ListIterator {
	return nil
}

func (String2Bytes) Length() int64 {
	return -1
}

func (String2Bytes) IsAbsent() bool {
	return false
}

func (String2Bytes) IsNull() bool {
	return false
}

func (String2Bytes) AsBool() (bool, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.AsBool()
}

func (String2Bytes) AsInt() (int64, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.AsInt()
}

func (String2Bytes) AsFloat() (float64, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.AsFloat()
}

func (n String2Bytes) AsString() (string, error) {
	return bytesToString(n.x), nil
}

func (n String2Bytes) AsBytes() ([]byte, error) {
	return n.x, nil
}

func (String2Bytes) AsLink() (datamodel.Link, error) {
	return mixins.String{TypeName: "dagjose.String2Bytes"}.AsLink()
}

func (String2Bytes) Prototype() datamodel.NodePrototype {
	return _String2Bytes__Prototype{}
}

type _String2Bytes__Prototype struct{}

func (_String2Bytes__Prototype) NewBuilder() datamodel.NodeBuilder {
	var nb _String2Bytes__Builder
	nb.Reset()
	return &nb
}

type _String2Bytes__Builder struct {
	_String2Bytes__Assembler
}

func (nb *_String2Bytes__Builder) Build() datamodel.Node {
	if *nb.m != schema.Maybe_Value {
		panic("invalid state: cannot call Build on an assembler that's not finished")
	}
	return nb.w
}

func (nb *_String2Bytes__Builder) Reset() {
	var w _String2Bytes
	var m schema.Maybe
	*nb = _String2Bytes__Builder{_String2Bytes__Assembler{w: &w, m: &m}}
}

type _String2Bytes__Assembler struct {
	w *_String2Bytes
	m *schema.Maybe
}

func (na *_String2Bytes__Assembler) reset() {}

func (_String2Bytes__Assembler) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.BeginMap(0)
}

func (_String2Bytes__Assembler) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.BeginList(0)
}

func (na *_String2Bytes__Assembler) AssignNull() error {
	switch *na.m {
	case allowNull:
		*na.m = schema.Maybe_Null
		return nil
	case schema.Maybe_Absent:
		return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.AssignNull()
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	panic("unreachable")
}

func (_String2Bytes__Assembler) AssignBool(bool) error {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.AssignBool(false)
}

func (_String2Bytes__Assembler) AssignInt(int64) error {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.AssignInt(0)
}

func (_String2Bytes__Assembler) AssignFloat(float64) error {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.AssignFloat(0)
}

func (na *_String2Bytes__Assembler) AssignString(v string) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	rawBytes, err := stringToBytes(v)
	if err != nil {
		return err
	}
	na.w.x = rawBytes
	*na.m = schema.Maybe_Value
	return nil
}

func (na *_String2Bytes__Assembler) AssignBytes(v []byte) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	na.w.x = v
	*na.m = schema.Maybe_Value
	return nil
}

func (_String2Bytes__Assembler) AssignLink(datamodel.Link) error {
	return mixins.StringAssembler{TypeName: "dagjose.String2Bytes"}.AssignLink(nil)
}

func (na *_String2Bytes__Assembler) AssignNode(v datamodel.Node) error {
	if v.IsNull() {
		return na.AssignNull()
	}
	if v2, ok := v.(*_String2Bytes); ok {
		switch *na.m {
		case schema.Maybe_Value, schema.Maybe_Null:
			panic("invalid state: cannot assign into assembler that's already finished")
		}
		*na.w = *v2
		*na.m = schema.Maybe_Value
		return nil
	}
	if v2, err := v.AsString(); err != nil {
		if v2, err := v.AsBytes(); err != nil {
			return err
		} else {
			return na.AssignBytes(v2)
		}
	} else {
		return na.AssignString(v2)
	}
}

func (_String2Bytes__Assembler) Prototype() datamodel.NodePrototype {
	return _String2Bytes__Prototype{}
}

func (String2Bytes) Type() schema.Type {
	return nil
}

func (n String2Bytes) Representation() datamodel.Node {
	return (*_String2Bytes__Repr)(n)
}

type _String2Bytes__Repr = _String2Bytes

var _ datamodel.Node = &_String2Bytes__Repr{}

type _String2Bytes__ReprPrototype = _String2Bytes__Prototype
type _String2Bytes__ReprAssembler = _String2Bytes__Assembler

func bytesToString(b []byte) string {
	// Remove the multibase prefix
	return base64.RawURLEncoding.EncodeToString(b)
}

func stringToBytes(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

//////////////////////////

// IgnoreMe matches the IPLD Schema type "IgnoreMe".  It has string kind.
type IgnoreMe = *_IgnoreMe
type _IgnoreMe struct{ x string }

func (n IgnoreMe) String() string {
	return n.x
}
func (_IgnoreMe__Prototype) fromString(w *_IgnoreMe, v string) error {
	*w = _IgnoreMe{v}
	return nil
}
func (_IgnoreMe__Prototype) FromString(v string) (IgnoreMe, error) {
	n := _IgnoreMe{v}
	return &n, nil
}

type _IgnoreMe__Maybe struct {
	m schema.Maybe
	v _IgnoreMe
}
type MaybeIgnoreMe = *_IgnoreMe__Maybe

func (m MaybeIgnoreMe) IsNull() bool {
	return m.m == schema.Maybe_Null
}
func (m MaybeIgnoreMe) IsAbsent() bool {
	return m.m == schema.Maybe_Absent
}
func (m MaybeIgnoreMe) Exists() bool {
	return m.m == schema.Maybe_Value
}
func (m MaybeIgnoreMe) AsNode() datamodel.Node {
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
func (m MaybeIgnoreMe) Must() IgnoreMe {
	if !m.Exists() {
		panic("unbox of a maybe rejected")
	}
	return &m.v
}

var _ datamodel.Node = (IgnoreMe)(&_IgnoreMe{})
var _ schema.TypedNode = (IgnoreMe)(&_IgnoreMe{})

func (IgnoreMe) Kind() datamodel.Kind {
	return datamodel.Kind_Link
}
func (IgnoreMe) LookupByString(string) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.LookupByString("")
}
func (IgnoreMe) LookupByNode(datamodel.Node) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.LookupByNode(nil)
}
func (IgnoreMe) LookupByIndex(idx int64) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.LookupByIndex(0)
}
func (IgnoreMe) LookupBySegment(seg datamodel.PathSegment) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.LookupBySegment(seg)
}
func (IgnoreMe) MapIterator() datamodel.MapIterator {
	return nil
}
func (IgnoreMe) ListIterator() datamodel.ListIterator {
	return nil
}
func (IgnoreMe) Length() int64 {
	return -1
}
func (IgnoreMe) IsAbsent() bool {
	return false
}
func (IgnoreMe) IsNull() bool {
	return false
}
func (IgnoreMe) AsBool() (bool, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.AsBool()
}
func (IgnoreMe) AsInt() (int64, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.AsInt()
}
func (IgnoreMe) AsFloat() (float64, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.AsFloat()
}
func (n IgnoreMe) AsString() (string, error) {
	return n.x, nil
}
func (IgnoreMe) AsBytes() ([]byte, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.AsBytes()
}
func (IgnoreMe) AsLink() (datamodel.Link, error) {
	return mixins.String{TypeName: "dagjose.IgnoreMe"}.AsLink()
}
func (IgnoreMe) Prototype() datamodel.NodePrototype {
	return _IgnoreMe__Prototype{}
}

type _IgnoreMe__Prototype struct{}

func (_IgnoreMe__Prototype) NewBuilder() datamodel.NodeBuilder {
	var nb _IgnoreMe__Builder
	nb.Reset()
	return &nb
}

type _IgnoreMe__Builder struct {
	_IgnoreMe__Assembler
}

func (nb *_IgnoreMe__Builder) Build() datamodel.Node {
	if *nb.m != schema.Maybe_Value {
		panic("invalid state: cannot call Build on an assembler that's not finished")
	}
	return nb.w
}
func (nb *_IgnoreMe__Builder) Reset() {
	var w _IgnoreMe
	var m schema.Maybe
	*nb = _IgnoreMe__Builder{_IgnoreMe__Assembler{w: &w, m: &m}}
}

type _IgnoreMe__Assembler struct {
	w *_IgnoreMe
	m *schema.Maybe
}

func (na *_IgnoreMe__Assembler) reset() {}
func (_IgnoreMe__Assembler) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.BeginMap(0)
}
func (_IgnoreMe__Assembler) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.BeginList(0)
}
func (na *_IgnoreMe__Assembler) AssignNull() error {
	switch *na.m {
	case allowNull:
		*na.m = schema.Maybe_Null
		return nil
	case schema.Maybe_Absent:
		return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignNull()
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	panic("unreachable")
}
func (_IgnoreMe__Assembler) AssignBool(bool) error {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignBool(false)
}
func (_IgnoreMe__Assembler) AssignInt(int64) error {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignInt(0)
}
func (_IgnoreMe__Assembler) AssignFloat(float64) error {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignFloat(0)
}
func (na *_IgnoreMe__Assembler) AssignString(v string) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	na.w.x = v
	*na.m = schema.Maybe_Value
	return nil
}
func (_IgnoreMe__Assembler) AssignBytes([]byte) error {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignBytes(nil)
}
func (_IgnoreMe__Assembler) AssignLink(datamodel.Link) error {
	return mixins.StringAssembler{TypeName: "dagjose.IgnoreMe"}.AssignLink(nil)
}
func (na *_IgnoreMe__Assembler) AssignNode(v datamodel.Node) error {
	if v.IsNull() {
		return na.AssignNull()
	}
	if v2, ok := v.(*_IgnoreMe); ok {
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
func (_IgnoreMe__Assembler) Prototype() datamodel.NodePrototype {
	return _IgnoreMe__Prototype{}
}
func (IgnoreMe) Type() schema.Type {
	return nil /*TODO:typelit*/
}
func (n IgnoreMe) Representation() datamodel.Node {
	return (*_IgnoreMe__Repr)(n)
}

type _IgnoreMe__Repr = _IgnoreMe

var _ datamodel.Node = &_IgnoreMe__Repr{}

type _IgnoreMe__ReprPrototype = _IgnoreMe__Prototype
type _IgnoreMe__ReprAssembler = _IgnoreMe__Assembler
