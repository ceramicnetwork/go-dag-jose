package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type jwsSignatureListAssembler struct{ d *DAGJOSE }

func (l *jwsSignatureListAssembler) AssembleValue() ipld.NodeAssembler {
	l.d.signatures = append(l.d.signatures, jwsSignature{})
	sigRef := &l.d.signatures[len(l.d.signatures)-1]
	return &jwsSignatureAssembler{
		signature: sigRef,
		key:       nil,
		state:     maStateInitial,
	}
}

func (l *jwsSignatureListAssembler) Finish() error {
	return nil
}

func (l *jwsSignatureListAssembler) ValuePrototype(_ int64) ipld.NodePrototype {
	return basicnode.Prototype.Map
}
