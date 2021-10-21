package dagjose

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/ipld/go-ipld-prime/schema"
)

var (
	_ ipld.Node          = dagJOSENode{}
	_ ipld.NodePrototype = &DagJOSENodePrototype{}
	_ ipld.NodeBuilder   = &dagJOSENodeBuilder{}
)

type DagJOSENodePrototype struct{}

func (d *DagJOSENodePrototype) NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJOSE: DagJOSE{}}
}

// NewBuilder Returns an instance of the DagJOSENodeBuilder which can be passed
// to ipld.Link.Load and will build a dagjose.DagJOSE object. This should only
// be necessary in reasonably advanced situations, most of the time you should
// be able to use dagjose.LoadJOSE.
func NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJOSE: DagJOSE{}}
}

type maState uint8

const (
	maState_initial     maState = iota // also the 'expect key or finish' state
	maState_midKey                     // waiting for a 'finished' state in the KeyAssembler
	maState_expectValue                // 'AssembleValue' is the only valid next step
	maState_midValue                   // waiting for a 'finished' state in the ValueAssembler
	maState_finished                   // finished
)

type ErrInvalidState struct {
	state maState
}

func (e ErrInvalidState) Error() string {
	return fmt.Sprintf("invalid state: %d", e.state)
}

// An implementation of `ipld.NodeBuilder` which builds a `dagjose.DagJOSE`
// object. This builder will throw an error if the IPLD data it is building
// does not match the schema specified in the spec
type dagJOSENodeBuilder struct {
	dagJOSE DagJOSE
	state   maState
	key     *string
}

var dagJOSEAssemblerMixin = mixins.MapAssembler{TypeName: "DagJOSEAssembler"}

func (d *dagJOSENodeBuilder) BeginMap(sizeHint int64) (ipld.MapAssembler, error) {
	if d.state != maState_initial {
		return nil, ErrInvalidState{d.state}
	}
	return d, nil
}

func (d *dagJOSENodeBuilder) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	if d.state != maState_midValue {
		return nil, ErrInvalidState{d.state}
	}
	if *d.key == "recipients" {
		d.dagJOSE.recipients = make([]jweRecipient, 0, sizeHint)
		d.state = maState_initial
		return &jweRecipientListAssembler{&d.dagJOSE}, nil
	}
	if *d.key == "signatures" {
		d.dagJOSE.signatures = make([]jwsSignature, 0, sizeHint)
		d.state = maState_initial
		return &jwsSignatureListAssembler{&d.dagJOSE}, nil
	}
	return dagJOSEAssemblerMixin.BeginList(sizeHint)
}

func (d *dagJOSENodeBuilder) AssignNull() error {
	if d.state != maState_midValue {
		return ErrInvalidState{d.state}
	}
	switch *d.key {
	case "payload":
		d.dagJOSE.payload = nil
	case "protected":
		d.dagJOSE.protected = nil
	case "unprotected":
		d.dagJOSE.unprotected = nil
	case "iv":
		d.dagJOSE.iv = nil
	case "aad":
		d.dagJOSE.aad = nil
	case "ciphertext":
		d.dagJOSE.ciphertext = nil
	case "tag":
		d.dagJOSE.tag = nil
	case "signatures":
		d.dagJOSE.signatures = nil
	case "recipients":
		d.dagJOSE.recipients = nil
	default:
		return dagJOSEAssemblerMixin.AssignNull()
	}
	d.state = maState_initial
	return nil
}

func (d *dagJOSENodeBuilder) AssignBool(b bool) error {
	return dagJOSEAssemblerMixin.AssignBool(b)
}

func (d *dagJOSENodeBuilder) AssignInt(i int64) error {
	return dagJOSEAssemblerMixin.AssignInt(i)
}

func (d *dagJOSENodeBuilder) AssignFloat(f float64) error {
	return dagJOSEAssemblerMixin.AssignFloat(f)
}

func (d *dagJOSENodeBuilder) AssignString(s string) error {
	if d.state != maState_midKey {
		return ErrInvalidState{d.state}
	}
	if !isValidJOSEKey(s) {
		return schema.ErrNoSuchField{Type: nil, Field: datamodel.PathSegmentOfString(s)}
	}
	d.key = &s
	d.state = maState_expectValue
	return nil
}

func (d *dagJOSENodeBuilder) AssignBytes(b []byte) error {
	if d.state != maState_midValue {
		return ErrInvalidState{d.state}
	}
	switch *d.key {
	case "payload":
		_, c, err := cid.CidFromBytes(b)
		if err != nil {
			return fmt.Errorf("payload is not a valid CID: %v", err)
		}
		d.dagJOSE.payload = &c
	case "protected":
		d.dagJOSE.protected = b
	case "unprotected":
		d.dagJOSE.unprotected = b
	case "iv":
		d.dagJOSE.iv = b
	case "aad":
		d.dagJOSE.aad = b
	case "ciphertext":
		d.dagJOSE.ciphertext = b
	case "tag":
		d.dagJOSE.tag = b
	case "signatures":
		return fmt.Errorf("attempted to assign bytes to 'signatures' key")
	case "recipients":
		return fmt.Errorf("attempted to assign bytes to 'recipients' key")
	default:
		return dagJOSEAssemblerMixin.AssignBytes(b)
	}
	d.state = maState_initial
	return nil
}

func (d *dagJOSENodeBuilder) AssignLink(l ipld.Link) error {
	return dagJOSEAssemblerMixin.AssignLink(l)
}
func (d *dagJOSENodeBuilder) AssignNode(n ipld.Node) error {
	return datamodel.Copy(n, d)
}
func (d *dagJOSENodeBuilder) Prototype() ipld.NodePrototype {
	return &DagJOSENodePrototype{}
}
func (d *dagJOSENodeBuilder) Build() ipld.Node {
	return dagJOSENode{&d.dagJOSE}
}
func (d *dagJOSENodeBuilder) Reset() {
}

func (d *dagJOSENodeBuilder) AssembleKey() ipld.NodeAssembler {
	if d.state != maState_initial {
		panic("misuse")
	}
	d.state = maState_midKey
	return d
}
func (d *dagJOSENodeBuilder) AssembleValue() ipld.NodeAssembler {
	if d.state != maState_expectValue {
		panic("misuse")
	}
	d.state = maState_midValue
	return d
}
func (d *dagJOSENodeBuilder) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if d.state != maState_initial {
		return nil, ErrInvalidState{d.state}
	}
	d.key = &k
	d.state = maState_midValue
	return d, nil
}
func (d *dagJOSENodeBuilder) Finish() error {
	if d.state != maState_initial {
		return ErrInvalidState{d.state}
	}
	d.state = maState_finished
	return nil
}
func (d *dagJOSENodeBuilder) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (d *dagJOSENodeBuilder) ValuePrototype(string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

func isValidJOSEKey(key string) bool {
	allowedKeys := []string{
		"payload",
		"signatures",
		"protected",
		"unprotected",
		"iv",
		"aad",
		"ciphertext",
		"tag",
		"recipients",
	}
	for _, allowedKey := range allowedKeys {
		if key == allowedKey {
			return true
		}
	}
	return false
}
