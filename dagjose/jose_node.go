package dagjose

import (
	"errors"
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipldBasicNode "github.com/ipld/go-ipld-prime/node/basic"
)

type dagJOSENode struct{ *DagJOSE }

func (d dagJOSENode) Kind() ipld.Kind {
	return ipld.Kind_Map
}
func (d dagJOSENode) LookupByString(key string) (ipld.Node, error) {
	if key == "payload" {
		if d.payload == nil {
			return keyOrNotFound(key, nil)
		}
		return ipldBasicNode.NewBytes(d.payload.Bytes()), nil
	}
	if key == "link" {
		if d.payload == nil {
			return keyOrNotFound(key, nil)
		}
		return ipldBasicNode.NewLink(cidlink.Link{Cid: *(d.payload)}), nil
	}
	if key == "signatures" {
		if d.signatures == nil {
			return keyOrNotFound(key, nil)
		}
		return &jwsSignaturesNode{d.signatures}, nil
	}
	if key == "protected" {
		return bytesOrNotFound(key, d.protected)
	}
	if key == "unprotected" {
		return bytesOrNotFound(key, d.unprotected)
	}
	if key == "iv" {
		return bytesOrNotFound(key, d.iv)
	}
	if key == "aad" {
		return bytesOrNotFound(key, d.aad)
	}
	if key == "ciphertext" {
		return bytesOrNotFound(key, d.ciphertext)
	}
	if key == "tag" {
		return bytesOrNotFound(key, d.tag)
	}
	if key == "recipients" {
		if d.recipients != nil {
			return fluent.MustBuildList(
				ipldBasicNode.Prototype.List,
				int64(len(d.recipients)),
				func(la fluent.ListAssembler) {
					for i := range d.recipients {
						la.AssembleValue().AssignNode(jweRecipientNode{&d.recipients[i]})
					}
				},
			), nil
		}
		return keyOrNotFound(key, nil)
	}
	return nil, fmt.Errorf("the key \"%v\" is not in dag-jose nodes", key)
}
func (d dagJOSENode) LookupByNode(key ipld.Node) (ipld.Node, error) {
	ks, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(ks)
}
func (d dagJOSENode) LookupByIndex(idx int64) (ipld.Node, error) {
	return nil, errors.New("can not lookup by index in of a map")
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
	return false, nil
}
func (d dagJOSENode) AsInt() (int64, error) {
	return 0, nil
}
func (d dagJOSENode) AsFloat() (float64, error) {
	return 0, nil
}
func (d dagJOSENode) AsString() (string, error) {
	return "", nil
}
func (d dagJOSENode) AsBytes() ([]byte, error) {
	return nil, nil
}
func (d dagJOSENode) AsLink() (ipld.Link, error) {
	return nil, nil
}
func (d dagJOSENode) Prototype() ipld.NodePrototype {
	return nil
}

// end ipld.Node implementation

func bytesOrNotFound(key string, b []byte) (ipld.Node, error) {
	return keyOrNotFound(key, ipldBasicNode.NewBytes(b))
}

func keyOrNotFound(key string, value ipld.Node) (ipld.Node, error) {
	if value != nil {
		return value, nil
	}
	return nil, fmt.Errorf("the key \"%v\" is not in this dag-jose node", key)
}

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
	value, _ := d.d.LookupByString(key)
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
		result = append(result, "link")
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
