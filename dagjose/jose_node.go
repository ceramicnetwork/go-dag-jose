package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	ipldBasicNode "github.com/ipld/go-ipld-prime/node/basic"
)

func (d *DagJOSE) ReprKind() ipld.ReprKind {
	return ipld.ReprKind_Map
}
func (d *DagJOSE) LookupByString(key string) (ipld.Node, error) {
	if key == "payload" {
		return ipldBasicNode.NewBytes(d.payload.Bytes()), nil
	}
	if key == "signatures" {
		return &joseSignaturesNode{d.signatures}, nil
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
				len(d.recipients),
				func(la fluent.ListAssembler) {
					for i := range d.recipients {
						la.AssembleValue().AssignNode(&d.recipients[i])
					}
				},
			), nil
		}
		return nil, nil
	}
	return nil, nil
}
func (d *DagJOSE) LookupByNode(key ipld.Node) (ipld.Node, error) {
	ks, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return d.LookupByString(ks)
}
func (d *DagJOSE) LookupByIndex(idx int) (ipld.Node, error) {
	return nil, nil
}
func (d *DagJOSE) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	return d.LookupByString(seg.String())
}
func (d *DagJOSE) MapIterator() ipld.MapIterator {
	return &DagJOSEMapIterator{
		d:     d,
		index: 0,
	}
}
func (d *DagJOSE) ListIterator() ipld.ListIterator {
	return nil
}
func (d *DagJOSE) Length() int {
	return len((&DagJOSEMapIterator{d: d, index: 0}).presentKeys())
}
func (d *DagJOSE) IsAbsent() bool {
	return false
}
func (d *DagJOSE) IsNull() bool {
	return false
}
func (d *DagJOSE) AsBool() (bool, error) {
	return false, nil
}
func (d *DagJOSE) AsInt() (int, error) {
	return 0, nil
}
func (d *DagJOSE) AsFloat() (float64, error) {
	return 0, nil
}
func (d *DagJOSE) AsString() (string, error) {
	return "", nil
}
func (d *DagJOSE) AsBytes() ([]byte, error) {
	return nil, nil
}
func (d *DagJOSE) AsLink() (ipld.Link, error) {
	return nil, nil
}
func (d *DagJOSE) Prototype() ipld.NodePrototype {
	return nil
}

// end ipld.Node implementation

func bytesOrNil(value []byte) ipld.Node {
	if value != nil {
		return ipldBasicNode.NewBytes(value)
	} else {
		return ipld.Absent
	}
}

type DagJOSEMapIterator struct {
	d     *DagJOSE
	index int
}

func (d *DagJOSEMapIterator) Next() (ipld.Node, ipld.Node, error) {
	if d.Done() {
		return nil, nil, ipld.ErrIteratorOverread{}
	}
	presentKeys := d.presentKeys()
	key := presentKeys[d.index]
	value, _ := d.d.LookupByString(key)
	d.index += 1
	return ipldBasicNode.NewString(key), value, nil
}

func (d *DagJOSEMapIterator) Done() bool {
	return d.index >= len(d.presentKeys())
}

func (d *DagJOSEMapIterator) presentKeys() []string {
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
