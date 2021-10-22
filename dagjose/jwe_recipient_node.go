package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type jweRecipientNode struct{ *jweRecipient }

var recipientNodeMixin = mixins.Map{TypeName: "JWERecipientNode"}

func (r jweRecipientNode) Kind() ipld.Kind {
	return ipld.Kind_Map
}
func (r jweRecipientNode) LookupByString(key string) (ipld.Node, error) {
	if key == "header" {
		return valueOrNotFound(
			key,
			r.header,
			func() (ipld.Node, error) {
				return fluent.MustBuildMap(
					basicnode.Prototype.Map,
					int64(len(r.header)),
					func(ma fluent.MapAssembler) {
						for key, value := range r.header {
							ma.AssembleEntry(key).AssignNode(value)
						}
					},
				), nil
			})
	}
	if key == "encrypted_key" {
		return valueOrNotFound(key, r.encrypted_key, nil)
	}
	return nil, nil
}
func (r jweRecipientNode) LookupByNode(key ipld.Node) (ipld.Node, error) {
	str, err := key.AsString()
	if err != nil {
		return nil, err
	}
	return r.LookupByString(str)
}
func (r jweRecipientNode) LookupByIndex(idx int64) (ipld.Node, error) {
	return recipientNodeMixin.LookupByIndex(idx)
}
func (r jweRecipientNode) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	return r.LookupByString(seg.String())
}
func (r jweRecipientNode) MapIterator() ipld.MapIterator {
	return &jweRecipientMapIterator{r: r, index: 0}
}
func (r jweRecipientNode) ListIterator() ipld.ListIterator {
	return nil
}
func (r jweRecipientNode) Length() int64 {
	if r.encrypted_key == nil && r.header == nil {
		return 0
	}
	if r.encrypted_key != nil && r.header != nil {
		return 2
	}
	return 1
}
func (r jweRecipientNode) IsAbsent() bool {
	return false
}
func (r jweRecipientNode) IsNull() bool {
	return false
}
func (r jweRecipientNode) AsBool() (bool, error) {
	return recipientNodeMixin.AsBool()
}
func (r jweRecipientNode) AsInt() (int64, error) {
	return recipientNodeMixin.AsInt()
}
func (r jweRecipientNode) AsFloat() (float64, error) {
	return recipientNodeMixin.AsFloat()
}
func (r jweRecipientNode) AsString() (string, error) {
	return recipientNodeMixin.AsString()
}
func (r jweRecipientNode) AsBytes() ([]byte, error) {
	return recipientNodeMixin.AsBytes()
}
func (r jweRecipientNode) AsLink() (ipld.Link, error) {
	return recipientNodeMixin.AsLink()
}
func (r jweRecipientNode) Prototype() ipld.NodePrototype {
	return nil
}

type jweRecipientMapIterator struct {
	r     jweRecipientNode
	index int
}

func (j *jweRecipientMapIterator) Next() (ipld.Node, ipld.Node, error) {
	if j.Done() {
		return nil, nil, ipld.ErrIteratorOverread{}
	}
	presentKeys := j.presentKeys()
	key := presentKeys[j.index]
	value, _ := j.r.LookupByString(key)
	j.index += 1
	return basicnode.NewString(key), value, nil
}

func (j *jweRecipientMapIterator) Done() bool {
	return j.index >= len(j.presentKeys())
}

func (j *jweRecipientMapIterator) presentKeys() []string {
	result := make([]string, 0)
	if j.r.header != nil {
		result = append(result, "header")
	}
	if j.r.encrypted_key != nil {
		result = append(result, "encrypted_key")
	}
	return result
}
