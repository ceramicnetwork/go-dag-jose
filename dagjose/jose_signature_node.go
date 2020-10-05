package dagjose

import (
	"strconv"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type joseSignaturesNode struct{ sigs []JWSSignature }

// joseSignatures Node implementation

func (d *joseSignaturesNode) ReprKind() ipld.ReprKind {
	return ipld.ReprKind_List
}
func (d *joseSignaturesNode) LookupByString(key string) (ipld.Node, error) {
	index, err := strconv.Atoi(key)
	if err != nil {
		return nil, nil
	}
	return d.LookupByIndex(index)
}
func (d *joseSignaturesNode) LookupByNode(key ipld.Node) (ipld.Node, error) {
	index, err := key.AsInt()
	if err != nil {
		return nil, err
	}
	return d.LookupByIndex(index)
}
func (d *joseSignaturesNode) LookupByIndex(idx int) (ipld.Node, error) {
	if len(d.sigs) > idx {
		return &d.sigs[idx], nil
	}
	return nil, nil
}
func (d *joseSignaturesNode) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	idx, err := seg.Index()
	if err != nil {
		return nil, nil
	}
	return d.LookupByIndex(idx)
}
func (d *joseSignaturesNode) MapIterator() ipld.MapIterator {
	return nil
}
func (d *joseSignaturesNode) ListIterator() ipld.ListIterator {
	return &joseSignaturesIterator{
		sigs:  d.sigs,
		index: 0,
	}
}
func (d *joseSignaturesNode) Length() int {
	return len(d.sigs)
}
func (d *joseSignaturesNode) IsAbsent() bool {
	return false
}
func (d *joseSignaturesNode) IsNull() bool {
	return false
}
func (d *joseSignaturesNode) AsBool() (bool, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsBool()
}
func (d *joseSignaturesNode) AsInt() (int, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsInt()
}
func (d *joseSignaturesNode) AsFloat() (float64, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsFloat()
}
func (d *joseSignaturesNode) AsString() (string, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsString()
}
func (d *joseSignaturesNode) AsBytes() ([]byte, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsBytes()
}
func (d *joseSignaturesNode) AsLink() (ipld.Link, error) {
	return mixins.List{TypeName: "jose.JOSESignature"}.AsLink()
}
func (d *joseSignaturesNode) Prototype() ipld.NodePrototype {
	return nil
}

// joseSignaturesNode ListIterator implementation

type joseSignaturesIterator struct {
	sigs  []JWSSignature
	index int
}

func (j *joseSignaturesIterator) Next() (idx int, value ipld.Node, err error) {
	if j.Done() {
		return 0, nil, ipld.ErrIteratorOverread{}
	}
	result := &j.sigs[j.index]
	j.index += 1
	return j.index, result, nil
}

func (j *joseSignaturesIterator) Done() bool {
	return j.index >= len(j.sigs)
}

// end ipld.Node implementation

// JOSESignature Node implementation

func (d *JWSSignature) ReprKind() ipld.ReprKind {
	return ipld.ReprKind_Map
}
func (d *JWSSignature) LookupByString(key string) (ipld.Node, error) {
	if key == "signature" {
		return basicnode.NewBytes(d.signature), nil
	}
	if key == "protected" {
		return basicnode.NewBytes(d.protected), nil
	}
	if key == "header" {
		if d.header == nil {
			return nil, nil
		}
		return fluent.MustBuildMap(
			basicnode.Prototype.Map,
			len(d.header),
			func(ma fluent.MapAssembler) {
				for key, value := range d.header {
					ma.AssembleEntry(key).AssignNode(value)
				}
			},
		), nil
	}
	return nil, nil
}
func (d *JWSSignature) LookupByNode(key ipld.Node) (ipld.Node, error) {
	keyString, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(keyString)
}
func (d *JWSSignature) LookupByIndex(idx int) (ipld.Node, error) {
	return nil, nil
}

func (d *JWSSignature) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	return d.LookupByString(seg.String())
}
func (d *JWSSignature) MapIterator() ipld.MapIterator {
	return &joseSignatureMapIterator{sig: d, index: 0}
}
func (d *JWSSignature) ListIterator() ipld.ListIterator {
	return nil
}
func (d *JWSSignature) Length() int {
	return len((&joseSignatureMapIterator{sig: d, index: 0}).presentKeys())
}
func (d *JWSSignature) IsAbsent() bool {
	return false
}
func (d *JWSSignature) IsNull() bool {
	return false
}
func (d *JWSSignature) AsBool() (bool, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsBool()
}
func (d *JWSSignature) AsInt() (int, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsInt()
}
func (d *JWSSignature) AsFloat() (float64, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsFloat()
}
func (d *JWSSignature) AsString() (string, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsString()
}
func (d *JWSSignature) AsBytes() ([]byte, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsBytes()
}
func (d *JWSSignature) AsLink() (ipld.Link, error) {
	return mixins.Map{TypeName: "dagjose.JOSESignature"}.AsLink()
}
func (d *JWSSignature) Prototype() ipld.NodePrototype {
	return nil
}

// end JOSESignature ipld.Node implementation

type joseSignatureMapIterator struct {
	sig   *JWSSignature
	index int
}

func (j *joseSignatureMapIterator) Next() (key ipld.Node, value ipld.Node, err error) {
	if j.Done() {
		return nil, nil, ipld.ErrIteratorOverread{}
	}
	keys := j.presentKeys()
	keyString := keys[j.index]
	value, _ = j.sig.LookupByString(keyString)
	j.index += 1
	return basicnode.NewString(keyString), value, nil
}

func (j *joseSignatureMapIterator) presentKeys() []string {
	result := []string{"signature"}
	if j.sig.protected != nil {
		result = append(result, "protected")
	}
	if j.sig.header != nil {
		result = append(result, "header")
	}
	return result
}

func (j *joseSignatureMapIterator) Done() bool {
	return j.index >= len(j.presentKeys())
}
