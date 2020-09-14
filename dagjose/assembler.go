package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

var (
	_ ipld.Node          = &DagJOSE{}
    _ ipld.NodePrototype = &DagJOSENodePrototype{}
	_ ipld.NodeAssembler = &DagJOSENodeBuilder{}
)

type DagJOSENodePrototype struct {}

func (d *DagJOSENodePrototype) NewBuilder() ipld.NodeBuilder {
    return &DagJOSENodeBuilder{dagJose: DagJOSE{}}
}

func NewBuilder() ipld.NodeBuilder {
    return &DagJOSENodeBuilder{dagJose: DagJOSE{}}
}


type maState uint8

const (
	maState_initial     maState = iota // also the 'expect key or finish' state
	maState_midKey                     // waiting for a 'finished' state in the KeyAssembler.
	maState_expectValue                // 'AssembleValue' is the only valid next step
	maState_midValue                   // waiting for a 'finished' state in the ValueAssembler.
	maState_finished                   // 'w' will also be nil, but this is a politer statement
)
type DagJOSENodeBuilder struct {
    dagJose DagJOSE
    state maState
	ka dagJoseKeyAssembler
	va dagJoseValueAssembler
}

type dagJoseKeyAssembler struct {*DagJOSENodeBuilder}
type dagJoseValueAssembler struct {*DagJOSENodeBuilder}

var dagJoseMixin = mixins.MapAssembler{TypeName: "dagjose"}

// Dummy node assembler implementation
// When complete this will actually be a generic way to construct a DagJOSE
// from any valid ipld codec. This feels a little redundant as dag-jose specifies
// that the wire protocol will be CBOR
func (d *DagJOSENodeBuilder) BeginMap(sizeHint int) (ipld.MapAssembler, error) {
    return &dagJOSEMapAssembler{d}, nil
}
func (d *DagJOSENodeBuilder) BeginList(sizeHint int) (ipld.ListAssembler, error) {
    return dagJoseMixin.BeginList(sizeHint)
}
func (d *DagJOSENodeBuilder) AssignNull() error {
    return dagJoseMixin.AssignNull()
}
func (d *DagJOSENodeBuilder) AssignBool(b bool) error {
    return dagJoseMixin.AssignBool(b)
}
func (d *DagJOSENodeBuilder) AssignInt(i int) error {
    return dagJoseMixin.AssignInt(i)
}
func (d *DagJOSENodeBuilder) AssignFloat(f float64) error {
    return dagJoseMixin.AssignFloat(f)
}
func (d *DagJOSENodeBuilder) AssignString(s string) error {
    return dagJoseMixin.AssignString(s)
}
func (d *DagJOSENodeBuilder) AssignBytes(b []byte) error {
    return dagJoseMixin.AssignBytes(b)
}
func (d *DagJOSENodeBuilder) AssignLink(l ipld.Link) error {
    return dagJoseMixin.AssignLink(l)
}
func (d *DagJOSENodeBuilder) AssignNode(n ipld.Node) error {
    // TODO
    return nil
}
func (d *DagJOSENodeBuilder) Prototype() ipld.NodePrototype {
    return &DagJOSENodePrototype{}
}
func (d *DagJOSENodeBuilder) Build() ipld.Node {
    return &d.dagJose
}
func (d *DagJOSENodeBuilder) Reset() {
}

type dagJOSEMapAssembler struct {*DagJOSENodeBuilder}

func (d *dagJOSEMapAssembler) AssembleKey() ipld.NodeAssembler {
    if d.state != maState_initial {
        panic("misuse")
    }
    return nil
}
func (d *dagJOSEMapAssembler) AssembleValue() ipld.NodeAssembler {
    return nil
}
func (d *dagJOSEMapAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
    return nil, nil
}
func (d *dagJOSEMapAssembler) Finish() error {
    return nil
}
func (d *dagJOSEMapAssembler) KeyPrototype() ipld.NodePrototype {
    return nil
}
func (d *dagJOSEMapAssembler) ValuePrototype(k string) ipld.NodePrototype {
    return nil
}

type dagJOSEHeaderAssembler struct {*DagJOSENodeBuilder}

func (ha *dagJOSEHeaderAssembler) AssembleKey() ipld.NodeAssembler {
    if ha.state != maState_initial {
        panic("misuse")
    }
    ha.state = maState_midKey
    return  nil
}

