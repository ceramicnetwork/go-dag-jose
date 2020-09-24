package dagjose

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type joseSignatureAssembler struct {
	signature *JOSESignature
	key       *string
	state     maState
}

var joseSignatureMixin = mixins.MapAssembler{TypeName: "JOSESignature"}

func (j *joseSignatureAssembler) BeginMap(sizeHint int) (ipld.MapAssembler, error) {
	if j.state == maState_midValue && *j.key == "header" {
		j.signature.header = make(map[string]string)
		j.state = maState_initial
		return &headerAssembler{
			header: j.signature.header,
			key:    nil,
			state:  maState_initial,
		}, nil
	}
	if j.state != maState_initial {
		panic("misuse")
	}
	return j, nil
}
func (j *joseSignatureAssembler) BeginList(sizeHint int) (ipld.ListAssembler, error) {
	return joseSignatureMixin.BeginList(sizeHint)
}
func (j *joseSignatureAssembler) AssignNull() error {
	if j.state == maState_midValue {
		switch *j.key {
		case "header":
			j.signature.header = nil
		case "protected":
			j.signature.protected = nil
		case "signature":
			j.signature.signature = nil
		default:
			panic("should never happen due to validation in map assembler")
		}
		return nil
	}
	return joseSignatureMixin.AssignNull()
}
func (j *joseSignatureAssembler) AssignBool(b bool) error {
	return joseSignatureMixin.AssignBool(b)
}
func (j *joseSignatureAssembler) AssignInt(i int) error {
	return joseSignatureMixin.AssignInt(i)
}
func (j *joseSignatureAssembler) AssignFloat(f float64) error {
	return joseSignatureMixin.AssignFloat(f)
}
func (j *joseSignatureAssembler) AssignString(s string) error {
	if j.state == maState_midKey {
		if !isValidJoseSignatureKey(s) {
			return fmt.Errorf("%s is not a vliad jose signature key", s)
		}
		j.key = &s
		j.state = maState_expectValue
		return nil
	}
	return joseSignatureMixin.AssignString(s)
}
func (j *joseSignatureAssembler) AssignBytes(b []byte) error {
	if j.state == maState_midValue {
		if *j.key == "protected" {
			j.signature.protected = b
			j.state = maState_initial
			return nil
		}
		if *j.key == "signature" {
			j.signature.signature = b
			j.state = maState_initial
			return nil
		}
		panic("should not be possible due to validation in map assembler")
	}
	return joseSignatureMixin.AssignBytes(b)
}
func (j *joseSignatureAssembler) AssignLink(l ipld.Link) error {
	return joseSignatureMixin.AssignLink(l)
}
func (j *joseSignatureAssembler) AssignNode(n ipld.Node) error {
	return fmt.Errorf("not implemented")
}
func (j *joseSignatureAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}

func (j *joseSignatureAssembler) AssembleKey() ipld.NodeAssembler {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_midKey
	return j
}

func (j *joseSignatureAssembler) AssembleValue() ipld.NodeAssembler {
	if j.state != maState_expectValue {
		panic("misuse")
	}
	j.state = maState_midValue
	return j
}
func (j *joseSignatureAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.key = &k
	j.state = maState_midValue
	return j, nil
}

func (j *joseSignatureAssembler) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (j *joseSignatureAssembler) ValuePrototype(k string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

func (j *joseSignatureAssembler) Finish() error {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_finished
	return nil
}

func isValidJoseSignatureKey(key string) bool {
	return key == "protected" || key == "header" || key == "signature"
}
