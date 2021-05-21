package dagjose

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

// This is a union of the DagJWE and DagJWS types. Typically you will want to
// use AsJWE and AsJWS to get a concrete JOSE object.
type DagJOSE struct {
	// JWS top level keys
	payload    *cid.Cid
	signatures []jwsSignature
	// JWE top level keys
	protected   []byte
	unprotected []byte
	iv          []byte
	aad         []byte
	ciphertext  []byte
	tag         []byte
	recipients  []jweRecipient
}

type jwsSignature struct {
	protected []byte
	header    map[string]ipld.Node
	signature []byte
}

type jweRecipient struct {
	header        map[string]ipld.Node
	encrypted_key []byte
}

func (d *DagJOSE) AsNode() ipld.Node {
	return dagJOSENode{d}
}

// If this jose object is a JWS then this will return a DagJWS, if it is a
// JWE then AsJWS will return nil
func (d *DagJOSE) AsJWS() *DagJWS {
	if d.payload != nil {
		return &DagJWS{dagjose: d}
	}
	return nil
}

// If this jose object is a JWE then this will return a DagJWE, if it is a
// JWS then AsJWE will return nil
func (d *DagJOSE) AsJWE() *DagJWE {
	if d.ciphertext != nil {
		return &DagJWE{dagjose: d}
	}
	return nil
}

type DagJWS struct{ dagjose *DagJOSE }

// Returns a DagJOSE object that implements ipld.Node and can be passed to
// ipld related infrastructure
func (d *DagJWS) AsJOSE() *DagJOSE {
	return d.dagjose
}

type DagJWE struct{ dagjose *DagJOSE }

// Returns a DagJOSE object that implements ipld.Node and can be passed to
// ipld related infrastructure
func (d *DagJWE) AsJOSE() *DagJOSE {
	return d.dagjose
}

func (d *DagJWS) PayloadLink() ipld.Link {
	return cidlink.Link{Cid: *d.dagjose.payload}
}
