package dagjose

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type joseRecipientAssembler struct {
	recipient *JWERecipient
	key       *string
	state     maState
}

var joseRecipientMixin = mixins.MapAssembler{TypeName: "JOSERecipient"}

func (j *joseRecipientAssembler) BeginMap(sizeHint int) (ipld.MapAssembler, error) {
	if j.state == maState_midValue && *j.key == "header" {
		j.recipient.header = make(map[string]ipld.Node)
		j.state = maState_initial
		return &headerAssembler{
			header: j.recipient.header,
			key:    nil,
			state:  maState_initial,
		}, nil
	}
	if j.state != maState_initial {
		panic("misuse")
	}
	return j, nil
}
func (j *joseRecipientAssembler) BeginList(sizeHint int) (ipld.ListAssembler, error) {
	return joseRecipientMixin.BeginList(sizeHint)
}
func (j *joseRecipientAssembler) AssignNull() error {
	if j.state == maState_midValue {
		switch *j.key {
		case "header":
			j.recipient.header = nil
		case "encrypted_key":
			j.recipient.encrypted_key = nil
		default:
			panic("should never happen due to validation in map assembler")
		}
		return nil
	}
	return joseRecipientMixin.AssignNull()
}
func (j *joseRecipientAssembler) AssignBool(b bool) error {
	return joseRecipientMixin.AssignBool(b)
}
func (j *joseRecipientAssembler) AssignInt(i int) error {
	return joseRecipientMixin.AssignInt(i)
}
func (j *joseRecipientAssembler) AssignFloat(f float64) error {
	return joseRecipientMixin.AssignFloat(f)
}
func (j *joseRecipientAssembler) AssignString(s string) error {
	if j.state == maState_midKey {
		if !isValidJoseRecipientKey(s) {
			return fmt.Errorf("%s is not a valid jose recipient key", s)
		}
		j.key = &s
		j.state = maState_expectValue
		return nil
	}
	return joseRecipientMixin.AssignString(s)
}
func (j *joseRecipientAssembler) AssignBytes(b []byte) error {
	if j.state == maState_midValue {
		if *j.key == "encrypted_key" {
			j.recipient.encrypted_key = b
			j.state = maState_initial
			return nil
		}
		panic("should not be possible due to validation in map assembler")
	}
	return joseRecipientMixin.AssignBytes(b)
}
func (j *joseRecipientAssembler) AssignLink(l ipld.Link) error {
	return joseRecipientMixin.AssignLink(l)
}
func (j *joseRecipientAssembler) AssignNode(n ipld.Node) error {
	return fmt.Errorf("not implemented")
}
func (j *joseRecipientAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}

func (j *joseRecipientAssembler) AssembleKey() ipld.NodeAssembler {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_midKey
	return j
}

func (j *joseRecipientAssembler) AssembleValue() ipld.NodeAssembler {
	if j.state != maState_expectValue {
		panic("misuse")
	}
	j.state = maState_midValue
	return j
}
func (j *joseRecipientAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.key = &k
	j.state = maState_midValue
	return j, nil
}

func (j *joseRecipientAssembler) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (j *joseRecipientAssembler) ValuePrototype(k string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

func (j *joseRecipientAssembler) Finish() error {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_finished
	return nil
}

func isValidJoseRecipientKey(key string) bool {
	return key == "encrypted_key" || key == "header"
}
