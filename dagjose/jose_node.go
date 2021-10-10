package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	ipldBasicNode "github.com/ipld/go-ipld-prime/node/basic"
)

type dagJOSENode struct{ DAGJOSE }

func (d dagJOSENode) Kind() ipld.Kind {
	return ipld.Kind_Map
}

func (d dagJOSENode) LookupByString(key string) (ipld.Node, error) {
	if key == "payload" {
		return ipldBasicNode.NewBytes(d.payload.Bytes()), nil
	}
	if key == "signatures" {
		return &jwsSignaturesNode{d.signatures}, nil
	}
	if key == "protected" {
		return bytesOrNil(d.protected), nil
	}
	if key == "unprotected" {
		return bytesOrNil(d.unprotected), nil
	}
	if key == "iv" {
		return bytesOrNil(d.iv), nil
	}
	if key == "aad" {
		return bytesOrNil(d.aad), nil
	}
	if key == "ciphertext" {
		return bytesOrNil(d.ciphertext), nil
	}
	if key == "tag" {
		return bytesOrNil(d.tag), nil
	}
	if key == "recipients" {
		if d.recipients != nil {
			return fluent.MustBuildList(
				ipldBasicNode.Prototype.List,
				int64(len(d.recipients)),
				func(la fluent.ListAssembler) {
					for i := range d.recipients {
						la.AssembleValue().AssignNode(jweRecipientNode{jweRecipient: &d.recipients[i]})
					}
				},
			), nil
		}
		return nil, nil
	}
	return nil, nil
}

func (d dagJOSENode) LookupByNode(key ipld.Node) (ipld.Node, error) {
	ks, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(ks)
}

func (d dagJOSENode) LookupByIndex(idx int64) (ipld.Node, error) {
	return nil, nil
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
	return new(dagJOSENodePrototype)
}

// end ipld.Node implementation

func bytesOrNil(value []byte) ipld.Node {
	if value != nil {
		return ipldBasicNode.NewBytes(value)
	} else {
		return ipld.Absent
	}
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
	var result []string
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
