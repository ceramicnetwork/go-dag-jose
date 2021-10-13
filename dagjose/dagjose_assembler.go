package dagjose

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

var (
	_ ipld.Node          = dagJOSENode{}
	_ ipld.NodePrototype = &DagJOSENodePrototype{}
	_ ipld.NodeBuilder   = &dagJOSENodeBuilder{}
)

type DagJOSENodePrototype struct{}

func (d *DagJOSENodePrototype) NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJose: DagJOSE{}}
}

// NewBuilder Returns an instance of the DagJOSENodeBuilder which can be passed
// to ipld.Link.Load and will build a dagjose.DagJOSE object. This should only
// be necessary in reasonably advanced situations, most of the time you should
// be able to use dagjose.LoadJOSE.
func NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJose: DagJOSE{}}
}

type maState uint8

const (
	maStateInitial     maState = iota // also the 'expect key or finish' state
	maStateMidKey                     // waiting for a 'finished' state in the KeyAssembler
	maStateExpectValue                // 'AssembleValue' is the only valid next step
	maStateMidValue                   // waiting for a 'finished' state in the ValueAssembler
	maStateFinished                   // finished
)

// An implementation of `ipld.NodeBuilder` which builds a `dagjose.DagJOSE`
// object. This builder will throw an error if the IPLD data it is building
// does not match the schema specified in the spec
type dagJOSENodeBuilder struct {
	dagJose DagJOSE
	state   maState
	key     *string
}

var dagJoseMixin = mixins.MapAssembler{TypeName: "dagJOSENodeAssembler"}

func (d *dagJOSENodeBuilder) BeginMap(sizeHint int64) (ipld.MapAssembler, error) {
	if d.state != maStateInitial {
		panic("misuse")
	}
	return d, nil
}
func (d *dagJOSENodeBuilder) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	if d.state == maStateMidValue && *d.key == "recipients" {
		d.dagJose.recipients = make([]jweRecipient, 0, sizeHint)
		d.state = maStateInitial
		return &jweRecipientListAssembler{&d.dagJose}, nil
	}
	if d.state == maStateMidValue && *d.key == "signatures" {
		d.dagJose.signatures = make([]jwsSignature, 0, sizeHint)
		d.state = maStateInitial
		return &jwsSignatureListAssembler{&d.dagJose}, nil
	}
	return dagJoseMixin.BeginList(sizeHint)
}
func (d *dagJOSENodeBuilder) AssignNull() error {
	if d.state == maStateMidValue {
		switch *d.key {
		case "payload":
			d.dagJose.payload = nil
		case "protected":
			d.dagJose.protected = nil
		case "unprotected":
			d.dagJose.unprotected = nil
		case "iv":
			d.dagJose.iv = nil
		case "aad":
			d.dagJose.aad = nil
		case "ciphertext":
			d.dagJose.ciphertext = nil
		case "tag":
			d.dagJose.tag = nil
		case "signatures":
			d.dagJose.signatures = nil
		case "recipients":
			d.dagJose.recipients = nil
		default:
			return dagJoseMixin.AssignNull()
		}
		d.state = maStateInitial
		return nil
	}
	return dagJoseMixin.AssignNull()
}
func (d *dagJOSENodeBuilder) AssignBool(b bool) error {
	return dagJoseMixin.AssignBool(b)
}
func (d *dagJOSENodeBuilder) AssignInt(i int64) error {
	return dagJoseMixin.AssignInt(i)
}
func (d *dagJOSENodeBuilder) AssignFloat(f float64) error {
	return dagJoseMixin.AssignFloat(f)
}
func (d *dagJOSENodeBuilder) AssignString(s string) error {
	if d.state == maStateMidKey {
		if !isValidJOSEKey(s) {
			return fmt.Errorf("attempted to assign an invalid JOSE key: %v", s)
		}
		d.key = &s
		d.state = maStateExpectValue
		return nil
	}
	return dagJoseMixin.AssignString(s)
}
func (d *dagJOSENodeBuilder) AssignBytes(b []byte) error {
	if d.state == maStateMidValue {
		switch *d.key {
		case "payload":
			_, c, err := cid.CidFromBytes(b)
			if err != nil {
				return fmt.Errorf("payload is not a valid CID: %v", err)
			}
			d.dagJose.payload = &c
		case "protected":
			d.dagJose.protected = b
		case "unprotected":
			d.dagJose.unprotected = b
		case "iv":
			d.dagJose.iv = b
		case "aad":
			d.dagJose.aad = b
		case "ciphertext":
			d.dagJose.ciphertext = b
		case "tag":
			d.dagJose.tag = b
		case "signatures":
			return fmt.Errorf("attempted to assign bytes to 'signatures' key")
		case "recipients":
			return fmt.Errorf("attempted to assign bytes to 'recipients' key")
		default:
			panic("should not happen due to AssignString implementation")
		}
		d.state = maStateInitial
		return nil
	}
	return dagJoseMixin.AssignBytes(b)
}
func (d *dagJOSENodeBuilder) AssignLink(l ipld.Link) error {
	return dagJoseMixin.AssignLink(l)
}
func (d *dagJOSENodeBuilder) AssignNode(n ipld.Node) error {
	return datamodel.Copy(n, d)
}
func (d *dagJOSENodeBuilder) Prototype() ipld.NodePrototype {
	return &DagJOSENodePrototype{}
}
func (d *dagJOSENodeBuilder) Build() ipld.Node {
	return dagJOSENode{&d.dagJose}
}
func (d *dagJOSENodeBuilder) Reset() {
}

func (d *dagJOSENodeBuilder) AssembleKey() ipld.NodeAssembler {
	if d.state != maStateInitial {
		panic("misuse")
	}
	d.state = maStateMidKey
	return d
}
func (d *dagJOSENodeBuilder) AssembleValue() ipld.NodeAssembler {
	if d.state != maStateExpectValue {
		panic("misuse")
	}
	d.state = maStateMidValue
	return d
}
func (d *dagJOSENodeBuilder) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if d.state != maStateInitial {
		panic("misuse")
	}
	d.key = &k
	d.state = maStateMidValue
	return d, nil
}
func (d *dagJOSENodeBuilder) Finish() error {
	if d.state != maStateInitial {
		panic("misuse")
	}
	d.state = maStateFinished
	return nil
}
func (d *dagJOSENodeBuilder) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (d *dagJOSENodeBuilder) ValuePrototype(k string) ipld.NodePrototype {
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
