package dagjose

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type headerAssembler struct {
	header map[string]string
	key    *string
	state  maState
}

func (h *headerAssembler) AssembleKey() ipld.NodeAssembler {
	if h.state != maState_initial {
		panic("misuse")
	}
	h.state = maState_midKey
	return h
}
func (h *headerAssembler) AssembleValue() ipld.NodeAssembler {
	if h.state != maState_expectValue {
		panic("misuse")
	}
	h.state = maState_midValue
	return h
}
func (h *headerAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if h.state != maState_initial {
		panic("misuse")
	}
	h.key = &k
	h.state = maState_midValue
	return h, nil
}
func (h *headerAssembler) Finish() error { return nil }
func (h *headerAssembler) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (h *headerAssembler) ValuePrototype(k string) ipld.NodePrototype {
	return basicnode.Prototype.String
}

var headerMixin = mixins.MapAssembler{TypeName: "header"}

func (h *headerAssembler) BeginMap(sizeHint int) (ipld.MapAssembler, error) {
	return mixins.StringAssembler{"string"}.BeginMap(0)
}
func (h *headerAssembler) BeginList(i int) (ipld.ListAssembler, error) {
	return headerMixin.BeginList(i)
}
func (h *headerAssembler) AssignNull() error {
	return headerMixin.AssignNull()
}
func (h *headerAssembler) AssignBool(b bool) error {
	return headerMixin.AssignBool(b)
}
func (h *headerAssembler) AssignInt(i int) error {
	return headerMixin.AssignInt(i)
}
func (h *headerAssembler) AssignFloat(f float64) error {
	return headerMixin.AssignFloat(f)
}
func (h *headerAssembler) AssignString(s string) error {
	if h.state == maState_midValue {
		h.header[*h.key] = s
		h.state = maState_initial
		return nil
	}
	if h.state == maState_midKey {
		h.key = &s
		h.state = maState_expectValue
		return nil
	}
	return headerMixin.AssignString(s)
}
func (h *headerAssembler) AssignBytes(b []byte) error {
	return headerMixin.AssignBytes(b)
}
func (h *headerAssembler) AssignLink(l ipld.Link) error {
	return headerMixin.AssignLink(l)
}
func (h *headerAssembler) AssignNode(n ipld.Node) error {
	if h.state == maState_midKey || h.state == maState_midValue {
		k, err := n.AsString()
		if err != nil {
			return fmt.Errorf("cannot assign non-string node into map key assembler")
		}
		return h.AssignString(k)
	}
	return fmt.Errorf("Attempted to assign node on header in bad state")
}
func (h *headerAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}
