package dagjose

import (
	"errors"
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type jweRecipientAssembler struct {
	recipient *jweRecipient
	key       *string
	state     maState
}

var jweRecipientMixin = mixins.MapAssembler{TypeName: "JOSERecipient"}

func (j *jweRecipientAssembler) BeginMap(_ int64) (ipld.MapAssembler, error) {
	if j.state == maStateMidValue && *j.key == "header" {
		j.recipient.header = make(map[string]ipld.Node)
		j.state = maStateInitial
		return &headerAssembler{
			header: j.recipient.header,
			key:    nil,
			state:  maStateInitial,
		}, nil
	}
	if j.state != maStateInitial {
		return nil, errors.New("misuse")
	}
	return j, nil
}

func (j *jweRecipientAssembler) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	return jweRecipientMixin.BeginList(sizeHint)
}

func (j *jweRecipientAssembler) AssignNull() error {
	if j.state == maStateMidValue {
		switch *j.key {
		case "header":
			j.recipient.header = nil
		case "encrypted_key":
			j.recipient.encryptedKey = nil
		default:
			return errors.New("should never happen due to validation in map assembler")
		}
		return nil
	}
	return jweRecipientMixin.AssignNull()
}

func (j *jweRecipientAssembler) AssignBool(b bool) error {
	return jweRecipientMixin.AssignBool(b)
}

func (j *jweRecipientAssembler) AssignInt(i int64) error {
	return jweRecipientMixin.AssignInt(i)
}

func (j *jweRecipientAssembler) AssignFloat(f float64) error {
	return jweRecipientMixin.AssignFloat(f)
}

func (j *jweRecipientAssembler) AssignString(s string) error {
	if j.state == maStateMidKey {
		if !isValidJWERecipientKey(s) {
			return fmt.Errorf("%s is not a valid JWE recipient key", s)
		}
		j.key = &s
		j.state = maStateExpectValue
		return nil
	}
	return jweRecipientMixin.AssignString(s)
}

func (j *jweRecipientAssembler) AssignBytes(b []byte) error {
	if j.state == maStateMidValue {
		if *j.key == "encrypted_key" {
			j.recipient.encryptedKey = b
			j.state = maStateInitial
			return nil
		}
		return errors.New("should not be possible due to validation in map assembler")
	}
	return jweRecipientMixin.AssignBytes(b)
}

func (j *jweRecipientAssembler) AssignLink(l ipld.Link) error {
	return jweRecipientMixin.AssignLink(l)
}

func (j *jweRecipientAssembler) AssignNode(_ ipld.Node) error {
	return fmt.Errorf("not implemented")
}

func (j *jweRecipientAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}

func (j *jweRecipientAssembler) AssembleKey() ipld.NodeAssembler {
	if j.state != maStateInitial {
		// TODO log err "misuse"
		return nil
	}
	j.state = maStateMidKey
	return j
}

func (j *jweRecipientAssembler) AssembleValue() ipld.NodeAssembler {
	if j.state != maStateExpectValue {
		// TODO log err "misuse"
		return nil
	}
	j.state = maStateMidValue
	return j
}

func (j *jweRecipientAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if j.state != maStateInitial {
		return nil, errors.New("misuse")
	}
	j.key = &k
	j.state = maStateMidValue
	return j, nil
}

func (j *jweRecipientAssembler) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}

func (j *jweRecipientAssembler) ValuePrototype(_ string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

func (j *jweRecipientAssembler) Finish() error {
	if j.state != maStateInitial {
		return errors.New("misuse")
	}
	j.state = maStateFinished
	return nil
}

func isValidJWERecipientKey(key string) bool {
	return key == "encrypted_key" || key == "header"
}
