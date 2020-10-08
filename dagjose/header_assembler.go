package dagjose

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type headerAssembler struct {
	header       map[string]ipld.Node
	valueBuilder ipld.NodeBuilder
	key          *string
	state        maState
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
	h.valueBuilder = basicnode.Prototype.Any.NewBuilder()
	return h.valueBuilder
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
	if h.state == maState_midValue {
		h.valueBuilder = basicnode.Prototype.Map.NewBuilder()
		ma, err := h.valueBuilder.BeginMap(sizeHint)
		if err != nil {
			return nil, err
		}
		hvam := headerValueAssemblerMap{
			ha: h,
			ma: ma,
		}
		return &hvam, nil
	}
	return mixins.StringAssembler{TypeName: "string"}.BeginMap(0)
}
func (h *headerAssembler) BeginList(sizeHint int) (ipld.ListAssembler, error) {
	if h.state == maState_midValue {
		h.state = maState_initial
		h.valueBuilder = basicnode.Prototype.List.NewBuilder()
		la, err := h.valueBuilder.BeginList(sizeHint)
		if err != nil {
			return nil, err
		}
		hval := headerValueAssemblerList{
			ha: h,
			la: la,
		}
		return &hval, nil
	}
	return headerMixin.BeginList(sizeHint)
}
func (h *headerAssembler) AssignNull() error {
	if h.state == maState_midValue {
		return h.AssignNode(ipld.Null)
	}
	return headerMixin.AssignNull()
}
func (h *headerAssembler) AssignBool(b bool) error {
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewBool(b))
	}
	return headerMixin.AssignBool(b)
}
func (h *headerAssembler) AssignInt(i int) error {
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewInt(i))
	}
	return headerMixin.AssignInt(i)
}
func (h *headerAssembler) AssignFloat(f float64) error {
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewFloat(f))
	}
	return headerMixin.AssignFloat(f)
}
func (h *headerAssembler) AssignString(s string) error {
	if h.state == maState_midKey {
		h.key = &s
		h.state = maState_expectValue
		return nil
	}
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewString(s))
	}
	return headerMixin.AssignString(s)
}
func (h *headerAssembler) AssignBytes(b []byte) error {
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewBytes(b))
	}
	return headerMixin.AssignBytes(b)
}
func (h *headerAssembler) AssignLink(l ipld.Link) error {
	if h.state == maState_midValue {
		return h.AssignNode(basicnode.NewLink(l))
	}
	return headerMixin.AssignLink(l)
}
func (h *headerAssembler) AssignNode(n ipld.Node) error {
	if h.state == maState_midKey {
		k, err := n.AsString()
		if err != nil {
			return fmt.Errorf("cannot get string from key: %v", err)
		}
		h.key = &k
		h.state = maState_expectValue
		return nil
	}
	if h.state == maState_midValue {
		h.header[*h.key] = n
		h.state = maState_initial
		h.key = nil
		h.valueBuilder = nil
		return nil
	}
	return fmt.Errorf("Attempted to assign node on header in bad state")
}
func (h *headerAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}

type headerValueAssemblerMap struct {
	ha *headerAssembler
	ma ipld.MapAssembler
}

func (hvam *headerValueAssemblerMap) AssembleKey() ipld.NodeAssembler {
	return hvam.ma.AssembleKey()
}

func (hvam *headerValueAssemblerMap) AssembleValue() ipld.NodeAssembler {
	return hvam.ma.AssembleValue()
}

func (hvam *headerValueAssemblerMap) AssembleEntry(s string) (ipld.NodeAssembler, error) {
	return hvam.ma.AssembleEntry(s)
}

func (hvam *headerValueAssemblerMap) Finish() error {
	if err := hvam.ma.Finish(); err != nil {
		return err
	}
	hvam.ha.header[*hvam.ha.key] = hvam.ha.valueBuilder.Build()
	hvam.ha.state = maState_initial
	hvam.ha.key = nil
	hvam.ha.valueBuilder = nil
	return nil
}

func (hvam *headerValueAssemblerMap) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}

func (hvam *headerValueAssemblerMap) ValuePrototype(k string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

type headerValueAssemblerList struct {
	ha *headerAssembler
	la ipld.ListAssembler
}

func (hval *headerValueAssemblerList) AssembleValue() ipld.NodeAssembler {
	return hval.la.AssembleValue()
}

func (hval *headerValueAssemblerList) Finish() error {
	if err := hval.la.Finish(); err != nil {
		return err
	}
	hval.ha.header[*hval.ha.key] = hval.ha.valueBuilder.Build()
	hval.ha.state = maState_initial
	hval.ha.key = nil
	hval.ha.valueBuilder = nil
	return nil
}

func (hval *headerValueAssemblerList) ValuePrototype(idx int) ipld.NodePrototype {
	return basicnode.Prototype.Any
}
