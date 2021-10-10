package dagjose

import (
	"fmt"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/pkg/errors"
)

type (
	dagJOSENodePrototype struct{}
	maState              uint
)

const (
	maStateInitial     maState = iota // also the 'expect key or finish' state
	maStateMidKey                     // waiting for a 'finished' state in the KeyAssembler.
	maStateExpectValue                // 'AssembleValue' is the only valid next step
	maStateMidValue                   // waiting for a 'finished' state in the ValueAssembler.
	maStateFinished                   // finished
)

var (
	dagJOSEMixin = mixins.MapAssembler{TypeName: "dagJOSE"}
)

func (d *dagJOSENodePrototype) NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJOSE: DAGJOSE{}}
}

// NewBuilder Returns an instance of the DagJOSENodeBuilder which can be passed to
// ipld.Link.Load and will build a dagjose.DAGJOSE object. This should only be
// necessary in reasonably advanced situations, most of the time you should be
// able to use dagjose.LoadJOSE
func NewBuilder() ipld.NodeBuilder {
	return &dagJOSENodeBuilder{dagJOSE: DAGJOSE{}}
}

// An implementation of `ipld.NodeBuilder` which builds a `dagjose.DAGJOSE`
// object. This builder will throw an error if the IPLD data it is building
// does not match the schema specified in the spec
type dagJOSENodeBuilder struct {
	dagJOSE DAGJOSE
	state   maState
	key     *string
}

func (d *dagJOSENodeBuilder) BeginMap(_ int64) (ipld.MapAssembler, error) {
	if d.state != maStateInitial {
		return nil, errors.New("misuse")
	}
	return d, nil
}
func (d *dagJOSENodeBuilder) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	if d.state == maStateMidValue && *d.key == "recipients" {
		d.dagJOSE.recipients = make([]jweRecipient, 0, sizeHint)
		d.state = maStateInitial
		return &jweRecipientListAssembler{d: &d.dagJOSE}, nil
	}
	if d.state == maStateMidValue && *d.key == "signatures" {
		d.dagJOSE.signatures = make([]jwsSignature, 0, sizeHint)
		d.state = maStateInitial
		return &jwsSignatureListAssembler{d: &d.dagJOSE}, nil
	}
	return dagJOSEMixin.BeginList(sizeHint)
}

func (d *dagJOSENodeBuilder) AssignNull() error {
	if d.state == maStateMidValue {
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
			return errors.New("should not happen due to AssignString implementation")
		}
		d.state = maStateInitial
		return nil
	}
	return dagJOSEMixin.AssignNull()
}

func (d *dagJOSENodeBuilder) AssignBool(b bool) error {
	return dagJOSEMixin.AssignBool(b)
}

func (d *dagJOSENodeBuilder) AssignInt(i int64) error {
	return dagJOSEMixin.AssignInt(i)
}

func (d *dagJOSENodeBuilder) AssignFloat(f float64) error {
	return dagJOSEMixin.AssignFloat(f)
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
	return dagJOSEMixin.AssignString(s)
}

func (d *dagJOSENodeBuilder) AssignBytes(b []byte) error {
	if d.state == maStateMidValue {
		switch *d.key {
		case "payload":
			_, id, err := cid.CidFromBytes(b)
			if err != nil {
				return errors.Wrap(err, "payload is not a valid CID: %v")
			}
			d.dagJOSE.payload = &id
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
			return errors.New("attempted to assign bytes to 'signatures' key")
		case "recipients":
			return errors.New("attempted to assign bytes to 'recipients' key")
		default:
			return errors.New("should not happen due to AssignString implementation")
		}
		d.state = maStateInitial
		return nil
	}
	return dagJOSEMixin.AssignBytes(b)
}

func (d *dagJOSENodeBuilder) AssignLink(l ipld.Link) error {
	return dagJOSEMixin.AssignLink(l)
}

func (d *dagJOSENodeBuilder) AssignNode(n ipld.Node) error {
	if d.state != maStateInitial {
		return errors.New("misuse")
	}
	if n.Kind() != ipld.Kind_Map {
		return ipld.ErrWrongKind{TypeName: "map", MethodName: "AssignNode", AppropriateKind: ipld.KindSet_JustMap, ActualKind: n.Kind()}
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

func (d *dagJOSENodeBuilder) Prototype() ipld.NodePrototype {
	return &dagJOSENodePrototype{}
}

func (d *dagJOSENodeBuilder) Build() ipld.Node {
	return dagJOSENode{DAGJOSE: d.dagJOSE}
}

func (d *dagJOSENodeBuilder) Reset() {}

func (d *dagJOSENodeBuilder) AssembleKey() ipld.NodeAssembler {
	if d.state != maStateInitial {
		// log error "misuse"
		return nil
	}
	d.state = maStateMidKey
	return d
}

func (d *dagJOSENodeBuilder) AssembleValue() ipld.NodeAssembler {
	if d.state != maStateExpectValue {
		// TODO log error
		return nil
	}
	d.state = maStateMidValue
	return d
}

func (d *dagJOSENodeBuilder) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if d.state != maStateInitial {
		return nil, errors.New("misuse")
	}
	d.key = &k
	d.state = maStateMidValue
	return d, nil
}

func (d *dagJOSENodeBuilder) Finish() error {
	if d.state != maStateInitial {
		return errors.New("misuse")
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
