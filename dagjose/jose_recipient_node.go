package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

func (r *JWERecipient) ReprKind() ipld.ReprKind {
    return ipld.ReprKind_Map
}
func (r *JWERecipient) LookupByString(key string) (ipld.Node, error) {
    if key == "header" {
        return fluent.MustBuildMap(
            basicnode.Prototype.Map,
            len(r.header),
            func(ma fluent.MapAssembler) {
                for key, value := range r.header {
                    ma.AssembleEntry(key).AssignString(value)
                }
            },
        ), nil
    }
    if key == "encrypted_key" {
        return stringOrNil(r.encrypted_key), nil
    }
    return nil, nil
}
func (r *JWERecipient) LookupByNode(key ipld.Node) (ipld.Node, error) {
    str, err := key.AsString()
    if err != nil {
        return nil, nil
    }
    return r.LookupByString(str)
}
func (r *JWERecipient) LookupByIndex(idx int) (ipld.Node, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.LookupByIndex(idx)
}
func (r *JWERecipient) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
    return r.LookupByString(seg.String())
}
func (r *JWERecipient) MapIterator() ipld.MapIterator {
    return &jweRecipientMapIterator{r: r, index: 0}
}
func (r *JWERecipient) ListIterator() ipld.ListIterator {
    return nil
}
func (r *JWERecipient) Length() int {
    if r.encrypted_key == nil && r.header == nil {
        return 0
    }
    if r.encrypted_key != nil && r.header != nil {
        return 2
    }
    return 1
}
func (r *JWERecipient) IsAbsent() bool {
    return false
}
func (r *JWERecipient) IsNull() bool {
    return false
}
func (r *JWERecipient) AsBool() (bool, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsBool()
}
func (r *JWERecipient) AsInt() (int, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsInt()
}
func (r *JWERecipient) AsFloat() (float64, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsFloat()
}
func (r *JWERecipient) AsString() (string, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsString()
}
func (r *JWERecipient) AsBytes() ([]byte, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsBytes()
}
func (r *JWERecipient) AsLink() (ipld.Link, error) {
    return mixins.Map{TypeName: "dagjose.JWERecipient"}.AsLink()
}
func (r *JWERecipient) Prototype() ipld.NodePrototype {
    return nil
}

type jweRecipientMapIterator struct {
    r *JWERecipient
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
