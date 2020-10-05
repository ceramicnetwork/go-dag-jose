package dagjose

import (
	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type joseSignatureListAssembler struct{ d *DagJOSE }

func (l *joseSignatureListAssembler) AssembleValue() ipld.NodeAssembler {
	l.d.signatures = append(l.d.signatures, JWSSignature{})
	sigRef := &l.d.signatures[len(l.d.signatures)-1]
	return &joseSignatureAssembler{
		signature: sigRef,
		key:       nil,
		state:     maState_initial,
	}
}

func (l *joseSignatureListAssembler) Finish() error {
	return nil
}
func (l *joseSignatureListAssembler) ValuePrototype(idx int) ipld.NodePrototype {
	return basicnode.Prototype.Map
}
