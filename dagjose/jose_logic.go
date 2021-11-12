package dagjose

import (
	"encoding/base64"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/ipld/go-ipld-prime/schema"
)

func (n Base64Url) String() string {
	return encodeBase64Url(n.x)
}

func (_Base64Url__Prototype) fromString(w *_Base64Url, v string) error {
	base64Url, err := Type.Base64Url.FromString(v)
	if err != nil {
		return err
	}
	*w = *base64Url
	return nil
}

func (_Base64Url__Prototype) FromString(v string) (Base64Url, error) {
	decoded, err := decodeBase64Url(v)
	if err != nil {
		return nil, err
	}
	return &_Base64Url{decoded}, nil
}

type _Base64Url__Maybe struct {
	m schema.Maybe
	v _Base64Url
}

type MaybeBase64Url = *_Base64Url__Maybe

func (m MaybeBase64Url) IsNull() bool {
	return m.m == schema.Maybe_Null
}
func (m MaybeBase64Url) IsAbsent() bool {
	return m.m == schema.Maybe_Absent
}
func (m MaybeBase64Url) Exists() bool {
	return m.m == schema.Maybe_Value
}
func (m MaybeBase64Url) AsNode() datamodel.Node {
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
func (m MaybeBase64Url) Must() Base64Url {
	if !m.Exists() {
		panic("unbox of a maybe rejected")
	}
	return &m.v
}

var _ datamodel.Node = (Base64Url)(&_Base64Url{})
var _ schema.TypedNode = (Base64Url)(&_Base64Url{})

func (Base64Url) Kind() datamodel.Kind {
	return datamodel.Kind_String
}
func (Base64Url) LookupByString(string) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.LookupByString("")
}
func (Base64Url) LookupByNode(datamodel.Node) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.LookupByNode(nil)
}
func (Base64Url) LookupByIndex(idx int64) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.LookupByIndex(0)
}
func (Base64Url) LookupBySegment(seg datamodel.PathSegment) (datamodel.Node, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.LookupBySegment(seg)
}
func (Base64Url) MapIterator() datamodel.MapIterator {
	return nil
}
func (Base64Url) ListIterator() datamodel.ListIterator {
	return nil
}
func (Base64Url) Length() int64 {
	return -1
}
func (Base64Url) IsAbsent() bool {
	return false
}
func (Base64Url) IsNull() bool {
	return false
}
func (Base64Url) AsBool() (bool, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.AsBool()
}
func (Base64Url) AsInt() (int64, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.AsInt()
}
func (Base64Url) AsFloat() (float64, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.AsFloat()
}
func (n Base64Url) AsString() (string, error) {
	return n.String(), nil
}
func (Base64Url) AsBytes() ([]byte, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.AsBytes()
}
func (Base64Url) AsLink() (datamodel.Link, error) {
	return mixins.String{TypeName: "dagjose.Base64Url"}.AsLink()
}
func (Base64Url) Prototype() datamodel.NodePrototype {
	return _Base64Url__Prototype{}
}

type _Base64Url__Prototype struct{}

func (_Base64Url__Prototype) NewBuilder() datamodel.NodeBuilder {
	var nb _Base64Url__Builder
	nb.Reset()
	return &nb
}

type _Base64Url__Builder struct {
	_Base64Url__Assembler
}

func (nb *_Base64Url__Builder) Build() datamodel.Node {
	if *nb.m != schema.Maybe_Value {
		panic("invalid state: cannot call Build on an assembler that's not finished")
	}
	return nb.w
}
func (nb *_Base64Url__Builder) Reset() {
	var w _Base64Url
	var m schema.Maybe
	*nb = _Base64Url__Builder{_Base64Url__Assembler{w: &w, m: &m}}
}

type _Base64Url__Assembler struct {
	w *_Base64Url
	m *schema.Maybe
}

func (na *_Base64Url__Assembler) reset() {}
func (_Base64Url__Assembler) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.BeginMap(0)
}
func (_Base64Url__Assembler) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.BeginList(0)
}
func (na *_Base64Url__Assembler) AssignNull() error {
	switch *na.m {
	case allowNull:
		*na.m = schema.Maybe_Null
		return nil
	case schema.Maybe_Absent:
		return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.AssignNull()
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	panic("unreachable")
}
func (_Base64Url__Assembler) AssignBool(bool) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.AssignBool(false)
}
func (_Base64Url__Assembler) AssignInt(int64) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.AssignInt(0)
}
func (_Base64Url__Assembler) AssignFloat(float64) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.AssignFloat(0)
}
func (na *_Base64Url__Assembler) AssignString(v string) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	decoded, err := decodeBase64Url(v)
	if err != nil {
		return err
	}
	na.w.x = decoded
	*na.m = schema.Maybe_Value
	return nil
}
func (na *_Base64Url__Assembler) AssignBytes(v []byte) error {
	switch *na.m {
	case schema.Maybe_Value, schema.Maybe_Null:
		panic("invalid state: cannot assign into assembler that's already finished")
	}
	na.w.x = string(v)
	*na.m = schema.Maybe_Value
	return nil
}
func (_Base64Url__Assembler) AssignLink(datamodel.Link) error {
	return mixins.StringAssembler{TypeName: "dagjose.Base64Url"}.AssignLink(nil)
}
func (na *_Base64Url__Assembler) AssignNode(v datamodel.Node) error {
	if v.IsNull() {
		return na.AssignNull()
	}
	if v2, ok := v.(*_Base64Url); ok {
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
func (_Base64Url__Assembler) Prototype() datamodel.NodePrototype {
	return _Base64Url__Prototype{}
}
func (Base64Url) Type() schema.Type {
	return nil /*TODO:typelit*/
}
func (n Base64Url) Representation() datamodel.Node {
	return (*_Base64Url__Repr)(n)
}

type _Base64Url__Repr = _Base64Url

var _ datamodel.Node = &_Base64Url__Repr{}

type _Base64Url__ReprPrototype = _Base64Url__Prototype
type _Base64Url__ReprAssembler = _Base64Url__Assembler

func (_Base64Url__Prototype) Cid(n Base64Url) (cid.Cid, error) {
	return cid.Cast([]byte(n.x))
}

func (_Base64Url__Prototype) Link(n Base64Url) (datamodel.Link, error) {
	c, err := Type.Base64Url.Cid(n)
	if err != nil {
		return nil, err
	}
	return cidlink.Link{Cid: c}, nil
}

func encodeBase64Url(decoded string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(decoded))
}

func decodeBase64Url(encoded string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	return string(decoded), err
}
