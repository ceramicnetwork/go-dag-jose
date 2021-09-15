package dagjose

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

// DagJOSE is a union of the DagJWE and DagJWS types. Typically, you will want
// to use AsJWE and AsJWS to get a concrete JOSE object.
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

// AsJWS will return a DagJWS if this jose object is a JWS, or nil if it is a
// JWE
func (d *DagJOSE) AsJWS() *DagJWS {
	if d.payload != nil {
		return &DagJWS{dagjose: d}
	}
	return nil
}

// AsJWE will return a DagJWE if this jose object is a JWE, or nil if it is a
// JWS
func (d *DagJOSE) AsJWE() *DagJWE {
	if d.ciphertext != nil {
		return &DagJWE{dagjose: d}
	}
	return nil
}

type DagJWS struct{ dagjose *DagJOSE }

// AsJOSE returns a DagJOSE object that implements ipld.Node and can be passed
// to ipld related infrastructure
func (d *DagJWS) AsJOSE() *DagJOSE {
	return d.dagjose
}

type DagJWE struct{ dagjose *DagJOSE }

// AsJOSE returns a DagJOSE object that implements ipld.Node and can be passed
// to ipld related infrastructure
func (d *DagJWE) AsJOSE() *DagJOSE {
	return d.dagjose
}

func (d *DagJWS) PayloadLink() ipld.Link {
	return cidlink.Link{Cid: *d.dagjose.payload}
}

// LinkPrototype will build CIDs using the dag-jose multicodec and the sha-384
// multihash
var LinkPrototype = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x85, // 0x85 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhLength: 48,   // sha3-224 hash has a 48-byte sum.
}}

// StoreJOSE is a convenience function which passes the correct
// dagjose.LinkPrototype append jose.AsNode() to ipld.LinkSystem.Store
func StoreJOSE(linkContext ipld.LinkContext, jose *DagJOSE, linkSystem ipld.LinkSystem) (ipld.Link, error) {
	return linkSystem.Store(linkContext, LinkPrototype, jose.AsNode())
}

var NodePrototype = &DagJOSENodePrototype{}

// LoadJOSE is a convenience function which wraps ipld.LinkSystem.Load. This
// will provide the dagjose.NodePrototype to the link system and attempt to
// cast the result to a DagJOSE object
func LoadJOSE(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (*DagJOSE, error) {
	n, err := linkSystem.Load(
		linkContext,
		lnk,
		NodePrototype,
	)
	if err != nil {
		return nil, err
	}

	return n.(dagJOSENode).DagJOSE, nil
}
