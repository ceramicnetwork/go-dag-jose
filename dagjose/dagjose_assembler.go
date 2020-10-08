package dagjose

import (
	"fmt"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

var (
	_ ipld.Node          = &DagJOSE{}
	_ ipld.NodePrototype = &DagJOSENodePrototype{}
	_ ipld.NodeAssembler = &DagJOSENodeBuilder{}
)

type DagJOSENodePrototype struct{}

func (d *DagJOSENodePrototype) NewBuilder() ipld.NodeBuilder {
	return &DagJOSENodeBuilder{dagJose: DagJOSE{}}
}

// Returns an instance of the DagJOSENodeBuilder which can be passed to
// ipld.Link.Load and will build a dagjose.DagJOSE object. This should only be
// neccesary in reasonably advanced situations, most of the time you should be
// able to use dagjose.LoadJOSE
func NewBuilder() ipld.NodeBuilder {
	return &DagJOSENodeBuilder{dagJose: DagJOSE{}}
}

type maState uint8

const (
	maState_initial     maState = iota // also the 'expect key or finish' state
	maState_midKey                     // waiting for a 'finished' state in the KeyAssembler.
	maState_expectValue                // 'AssembleValue' is the only valid next step
	maState_midValue                   // waiting for a 'finished' state in the ValueAssembler.
	maState_finished                   // finised
)

// An implementation of `ipld.NodeBuilder` which builds a `dagjose.DagJOSE`
// object. This builder will throw an error if the IPLD data it is building
// does not match the schema specified in the spec
type DagJOSENodeBuilder struct {
	dagJose DagJOSE
	state   maState
	key     *string
}

var dagJoseMixin = mixins.MapAssembler{TypeName: "dagjose"}

func (d *DagJOSENodeBuilder) BeginMap(sizeHint int) (ipld.MapAssembler, error) {
	if d.state != maState_initial {
		panic("misuse")
	}
	return d, nil
}
func (d *DagJOSENodeBuilder) BeginList(sizeHint int) (ipld.ListAssembler, error) {
	if d.state == maState_midValue && *d.key == "recipients" {
		d.dagJose.recipients = make([]JWERecipient, 0, sizeHint)
		d.state = maState_initial
		return &joseRecipientListAssembler{&d.dagJose}, nil
	}
	if d.state == maState_midValue && *d.key == "signatures" {
		d.dagJose.signatures = make([]JWSSignature, 0, sizeHint)
		d.state = maState_initial
		return &joseSignatureListAssembler{&d.dagJose}, nil
	}
	return dagJoseMixin.BeginList(sizeHint)
}
func (d *DagJOSENodeBuilder) AssignNull() error {
	if d.state == maState_midValue {
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
			panic("should not happen due to AssignString implementation")
		}
		d.state = maState_initial
		return nil
	}
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
	if d.state == maState_midKey {
		if !isValidJOSEKey(s) {
			return fmt.Errorf("Attempted to assign an invalid JOSE key: %v", s)
		}
		d.key = &s
		d.state = maState_expectValue
		return nil
	}
	return dagJoseMixin.AssignString(s)
}
func (d *DagJOSENodeBuilder) AssignBytes(b []byte) error {
	if d.state == maState_midValue {
		switch *d.key {
		case "payload":
            _, cid, err := cid.CidFromBytes(b)
            if err != nil {
                return fmt.Errorf("payload is not a valid CID: %v", err)
            }
			d.dagJose.payload = &cid
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
		d.state = maState_initial
		return nil
	}
	return dagJoseMixin.AssignBytes(b)
}
func (d *DagJOSENodeBuilder) AssignLink(l ipld.Link) error {
	return dagJoseMixin.AssignLink(l)
}
func (d *DagJOSENodeBuilder) AssignNode(n ipld.Node) error {
	if d.state != maState_initial {
		panic("misuse")
	}
	if n.ReprKind() != ipld.ReprKind_Map {
		return ipld.ErrWrongKind{TypeName: "map", MethodName: "AssignNode", AppropriateKind: ipld.ReprKindSet_JustMap, ActualKind: n.ReprKind()}
	}
	itr := n.MapIterator()
	for !itr.Done() {
		k, v, err := itr.Next()
		if err != nil {
			return err
		}
		if err := d.AssembleKey().AssignNode(k); err != nil {
			return err
		}
		if err := d.AssembleValue().AssignNode(v); err != nil {
			return err
		}
	}
	return d.Finish()
}
func (d *DagJOSENodeBuilder) Prototype() ipld.NodePrototype {
	return &DagJOSENodePrototype{}
}
func (d *DagJOSENodeBuilder) Build() ipld.Node {
	return &d.dagJose
}
func (d *DagJOSENodeBuilder) Reset() {
}

func (d *DagJOSENodeBuilder) AssembleKey() ipld.NodeAssembler {
	if d.state != maState_initial {
		panic("misuse")
	}
	d.state = maState_midKey
	return d
}
func (d *DagJOSENodeBuilder) AssembleValue() ipld.NodeAssembler {
	if d.state != maState_expectValue {
		panic("misuse")
	}
	d.state = maState_midValue
	return d
}
func (d *DagJOSENodeBuilder) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if d.state != maState_initial {
		panic("misuse")
	}
	d.key = &k
	d.state = maState_midValue
	return d, nil
}
func (d *DagJOSENodeBuilder) Finish() error {
	if d.state != maState_initial {
		panic("misuse")
	}
	d.state = maState_finished
	return nil
}
func (d *DagJOSENodeBuilder) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (d *DagJOSENodeBuilder) ValuePrototype(k string) ipld.NodePrototype {
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
