package dagjose

import (
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/basic"
)

type jwsSignatureListAssembler struct{ d *DagJOSE }

func (l *jwsSignatureListAssembler) AssembleValue() ipld.NodeAssembler {
	l.d.signatures = append(l.d.signatures, jwsSignature{})
	sigRef := &l.d.signatures[len(l.d.signatures)-1]
	return &jwsSignatureAssembler{
		signature: sigRef,
		key:       nil,
		state:     maState_initial,
	}
}

func (l *jwsSignatureListAssembler) Finish() error {
	return nil
}
func (l *jwsSignatureListAssembler) ValuePrototype(idx int64) ipld.NodePrototype {
	return basicnode.Prototype.Map
}
