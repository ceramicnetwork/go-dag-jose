package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/linking/cid"
	ipldBasicNode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type dagJOSENode struct{ *DagJOSE }

var dagJOSEMixin = mixins.Map{TypeName: "dagJOSE"}

func (d dagJOSENode) Kind() ipld.Kind {
	return ipld.Kind_Map
}
func (d dagJOSENode) LookupByString(key string) (ipld.Node, error) {
	if key == "payload" {
		return valueOrNotFound(
			key,
			d.payload,
			func() (ipld.Node, error) {
				return ipldBasicNode.NewBytes(d.payload.Bytes()), nil
			},
		)
	}
	if key == "link" {
		return valueOrNotFound(
			key,
			d.payload,
			func() (ipld.Node, error) {
				return ipldBasicNode.NewLink(cidlink.Link{Cid: *(d.payload)}), nil
			},
		)
	}
	if key == "signatures" {
		return valueOrNotFound(
			key,
			d.signatures,
			func() (ipld.Node, error) {
				return &jwsSignaturesNode{d.signatures}, nil
			},
		)
	}
	if key == "protected" {
		return valueOrNotFound(key, d.protected, nil)
	}
	if key == "unprotected" {
		return valueOrNotFound(key, d.unprotected, nil)
	}
	if key == "iv" {
		return valueOrNotFound(key, d.iv, nil)
	}
	if key == "aad" {
		return valueOrNotFound(key, d.aad, nil)
	}
	if key == "ciphertext" {
		return valueOrNotFound(key, d.ciphertext, nil)
	}
	if key == "tag" {
		return valueOrNotFound(key, d.tag, nil)
	}
	if key == "recipients" {
		return valueOrNotFound(
			key,
			d.recipients,
			func() (ipld.Node, error) {
				return fluent.MustBuildList(
					ipldBasicNode.Prototype.List,
					int64(len(d.recipients)),
					func(la fluent.ListAssembler) {
						for i := range d.recipients {
							la.AssembleValue().AssignNode(jweRecipientNode{&d.recipients[i]})
						}
					},
				), nil
			},
		)
	}
	return valueOrNotFound(key, nil, nil)
}
func (d dagJOSENode) LookupByNode(key ipld.Node) (ipld.Node, error) {
	ks, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(ks)
}
func (d dagJOSENode) LookupByIndex(idx int64) (ipld.Node, error) {
	return dagJOSEMixin.LookupByIndex(idx)
}
func (d dagJOSENode) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	return d.LookupByString(seg.String())
}
func (d dagJOSENode) MapIterator() ipld.MapIterator {
	return &dagJOSEMapIterator{
		d:     d,
		index: 0,
	}
}
func (d dagJOSENode) ListIterator() ipld.ListIterator {
	return nil
}
func (d dagJOSENode) Length() int64 {
	return int64(len((&dagJOSEMapIterator{d: d, index: 0}).presentKeys()))
}
func (d dagJOSENode) IsAbsent() bool {
	return false
}
func (d dagJOSENode) IsNull() bool {
	return false
}
func (d dagJOSENode) AsBool() (bool, error) {
	return dagJOSEMixin.AsBool()
}
func (d dagJOSENode) AsInt() (int64, error) {
	return dagJOSEMixin.AsInt()
}
func (d dagJOSENode) AsFloat() (float64, error) {
	return dagJOSEMixin.AsFloat()
}
func (d dagJOSENode) AsString() (string, error) {
	return dagJOSEMixin.AsString()
}
func (d dagJOSENode) AsBytes() ([]byte, error) {
	return dagJOSEMixin.AsBytes()
}
func (d dagJOSENode) AsLink() (ipld.Link, error) {
	return dagJOSEMixin.AsLink()
}
func (d dagJOSENode) Prototype() ipld.NodePrototype {
	return &DagJOSENodePrototype{}
}

// end ipld.Node implementation

type dagJOSEMapIterator struct {
	d     dagJOSENode
	index int
}

func (d *dagJOSEMapIterator) Next() (ipld.Node, ipld.Node, error) {
	if d.Done() {
		return nil, nil, ipld.ErrIteratorOverread{}
	}
	presentKeys := d.presentKeys()
	key := presentKeys[d.index]
	value, err := d.d.LookupByString(key)
	if err != nil {
		return nil, nil, err
	}
	d.index += 1
	return ipldBasicNode.NewString(key), value, nil
}

func (d *dagJOSEMapIterator) Done() bool {
	return d.index >= len(d.presentKeys())
}

func (d *dagJOSEMapIterator) presentKeys() []string {
	result := make([]string, 0)
	if d.d.payload != nil {
		result = append(result, "payload")
	}
	if d.d.signatures != nil {
		result = append(result, "signatures")
	}
	if d.d.protected != nil {
		result = append(result, "protected")
	}
	if d.d.unprotected != nil {
		result = append(result, "unprotected")
	}
	if d.d.iv != nil {
		result = append(result, "iv")
	}
	if d.d.aad != nil {
		result = append(result, "aad")
	}
	if d.d.ciphertext != nil {
		result = append(result, "ciphertext")
	}
	if d.d.tag != nil {
		result = append(result, "tag")
	}
	if d.d.recipients != nil {
		result = append(result, "recipients")
	}
	return result
}
