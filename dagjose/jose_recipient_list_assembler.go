package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type joseRecipientListAssembler struct{ d *DagJOSE }

func (l *joseRecipientListAssembler) AssembleValue() ipld.NodeAssembler {
	l.d.recipients = append(l.d.recipients, JWERecipient{})
	nextRef := &l.d.recipients[len(l.d.recipients)-1]
	return &joseRecipientAssembler{
		recipient: nextRef,
		key:       nil,
		state:     maState_initial,
	}
}

func (l *joseRecipientListAssembler) Finish() error {
	return nil
}
func (l *joseRecipientListAssembler) ValuePrototype(idx int) ipld.NodePrototype {
	return basicnode.Prototype.Map
}
