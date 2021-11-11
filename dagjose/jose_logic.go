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

// IgnoreLink matches the IPLD Schema type "IgnoreLink".  It has link kind.
type IgnoreLink = *_IgnoreLink

// TODO
type _IgnoreLink struct{ x datamodel.Link }

func (n IgnoreLink) Link() datamodel.Link {
	return n.x
}
func (_IgnoreLink__Prototype) FromLink(v datamodel.Link) (IgnoreLink, error) {
	n := _IgnoreLink{v}
	return &n, nil
}

type _IgnoreLink__Maybe struct {
	m schema.Maybe
	v _IgnoreLink
}
type MaybeIgnoreLink = *_IgnoreLink__Maybe

func (m MaybeIgnoreLink) IsNull() bool {
	return m.m == schema.Maybe_Null
}
func (m MaybeIgnoreLink) IsAbsent() bool {
	return m.m == schema.Maybe_Absent
}
func (m MaybeIgnoreLink) Exists() bool {
	return m.m == schema.Maybe_Value
}
func (m MaybeIgnoreLink) AsNode() datamodel.Node {
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
func (m MaybeIgnoreLink) Must() IgnoreLink {
	if !m.Exists() {
		panic("unbox of a maybe rejected")
	}
	return &m.v
}

var _ datamodel.Node = (IgnoreLink)(&_IgnoreLink{})
var _ schema.TypedNode = (IgnoreLink)(&_IgnoreLink{})

func (IgnoreLink) Kind() datamodel.Kind {
	return datamodel.Kind_Link
}
func (IgnoreLink) LookupByString(string) (datamodel.Node, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.LookupByString("")
}
func (IgnoreLink) LookupByNode(datamodel.Node) (datamodel.Node, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.LookupByNode(nil)
}
func (IgnoreLink) LookupByIndex(idx int64) (datamodel.Node, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.LookupByIndex(0)
}
func (IgnoreLink) LookupBySegment(seg datamodel.PathSegment) (datamodel.Node, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.LookupBySegment(seg)
}
func (IgnoreLink) MapIterator() datamodel.MapIterator {
	return nil
}
func (IgnoreLink) ListIterator() datamodel.ListIterator {
	return nil
}
func (IgnoreLink) Length() int64 {
	return -1
}
func (IgnoreLink) IsAbsent() bool {
	return false
}
func (IgnoreLink) IsNull() bool {
	return false
}
func (IgnoreLink) AsBool() (bool, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.AsBool()
}
func (IgnoreLink) AsInt() (int64, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.AsInt()
}
func (IgnoreLink) AsFloat() (float64, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.AsFloat()
}
func (IgnoreLink) AsString() (string, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.AsString()
}
func (IgnoreLink) AsBytes() ([]byte, error) {
	return mixins.Link{TypeName: "dagjose.IgnoreLink"}.AsBytes()
}
func (n IgnoreLink) AsLink() (datamodel.Link, error) {
	return n.x, nil
}
func (IgnoreLink) Prototype() datamodel.NodePrototype {
	return _IgnoreLink__Prototype{}
}

type _IgnoreLink__Prototype struct{}

func (_IgnoreLink__Prototype) NewBuilder() datamodel.NodeBuilder {
	var nb _IgnoreLink__Builder
	nb.Reset()
	return &nb
}

type _IgnoreLink__Builder struct {
	_IgnoreLink__Assembler
}

func (nb *_IgnoreLink__Builder) Build() datamodel.Node {
	if *nb.m != schema.Maybe_Value {
		panic("invalid state: cannot call Build on an assembler that's not finished")
	}
	return nb.w
}
func (nb *_IgnoreLink__Builder) Reset() {
	var w _IgnoreLink
	var m schema.Maybe
	*nb = _IgnoreLink__Builder{_IgnoreLink__Assembler{w: &w, m: &m}}
}

type _IgnoreLink__Assembler struct {
	w *_IgnoreLink
	m *schema.Maybe
}

func (na *_IgnoreLink__Assembler) reset() {}
func (_IgnoreLink__Assembler) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.BeginMap(0)
}
func (_IgnoreLink__Assembler) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.BeginList(0)
}
func (na *_IgnoreLink__Assembler) AssignNull() error {
	switch *na.m {
	case allowNull:
		*na.m = schema.Maybe_Null
		return nil
	case schema.Maybe_Absent:
		return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignNull()
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	panic("unreachable")
}
func (_IgnoreLink__Assembler) AssignBool(bool) error {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignBool(false)
}
func (_IgnoreLink__Assembler) AssignInt(int64) error {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignInt(0)
}
func (_IgnoreLink__Assembler) AssignFloat(float64) error {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignFloat(0)
}
func (_IgnoreLink__Assembler) AssignString(string) error {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignString("")
}
func (_IgnoreLink__Assembler) AssignBytes([]byte) error {
	return mixins.LinkAssembler{TypeName: "dagjose.IgnoreLink"}.AssignBytes(nil)
}
func (na *_IgnoreLink__Assembler) AssignLink(v datamodel.Link) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	na.w.x = v
	*na.m = schema.Maybe_Value
	return nil
}
func (na *_IgnoreLink__Assembler) AssignNode(v datamodel.Node) error {
	if v.IsNull() {
		return na.AssignNull()
	}
	if v2, ok := v.(*_IgnoreLink); ok {
		switch *na.m {
		case schema.Maybe_Value, schema.Maybe_Null:
			panic("invalid state: cannot assign into assembler that's already finished")
		}
		*na.w = *v2
		*na.m = schema.Maybe_Value
		return nil
	}
	if v2, err := v.AsLink(); err != nil {
		return err
	} else {
		return na.AssignLink(v2)
	}
}
func (_IgnoreLink__Assembler) Prototype() datamodel.NodePrototype {
	return _IgnoreLink__Prototype{}
}
func (IgnoreLink) Type() schema.Type {
	return nil /*TODO:typelit*/
}
func (n IgnoreLink) Representation() datamodel.Node {
	return (*_IgnoreLink__Repr)(n)
}

type _IgnoreLink__Repr = _IgnoreLink

var _ datamodel.Node = &_IgnoreLink__Repr{}

type _IgnoreLink__ReprPrototype = _IgnoreLink__Prototype
type _IgnoreLink__ReprAssembler = _IgnoreLink__Assembler
