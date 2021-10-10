package dagjose

import (
	"errors"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

// DAGJOSE is a union of the DAGJWE and DAGJWS types. Typically, you will want to
// use AsJWE and AsJWS to get a concrete JOSE object.
type DAGJOSE struct {
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
	header       map[string]ipld.Node
	encryptedKey []byte
}

func (d DAGJOSE) AsNode() ipld.Node {
	return dagJOSENode{DAGJOSE: d}
}

// AsJWS If this JOSE object is a JWS then this will return a DAGJWS, if it is a
// JWE then AsJWS will return nil
func (d DAGJOSE) AsJWS() *DAGJWS {
	if d.payload != nil {
		return &DAGJWS{dagJOSE: d}
	}
	return nil
}

// AsJWE If this jose object is a JWE then this will return a DAGJWE, if it is a
// JWS then AsJWE will return nil
func (d DAGJOSE) AsJWE() *DAGJWE {
	if d.ciphertext != nil {
		return &DAGJWE{dagJOSE: d}
	}
	return nil
}

type DAGJWS struct{ dagJOSE DAGJOSE }

// AsJOSE Returns a DAGJOSE object that implements ipld.Node and can be passed to
// IPLD related infrastructure
func (d DAGJWS) AsJOSE() *DAGJOSE {
	return &d.dagJOSE
}

type DAGJWE struct{ dagJOSE DAGJOSE }

// AsJOSE Returns a DAGJOSE object that implements ipld.Node and can be passed to
// ipld related infrastructure
func (d DAGJWE) AsJOSE() *DAGJOSE {
	return &d.dagJOSE
}

func (d DAGJWS) PayloadLink() ipld.Link {
	return cidlink.Link{Cid: *d.dagJOSE.payload}
}

// LinkPrototype A link prototype which will build CIDs using the dag-jose multicodec and
// the sha-384 multihash
var LinkPrototype = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x85, // 0x71 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhLength: 48,   // sha3-224 hash has a 48-byte sum.
}}

// StoreJOSE A convenience function which passes the correct dagJOSE.LinkProtoype append
// jose.AsNode() to ipld.LinkSystem.Store
func StoreJOSE(linkContext ipld.LinkContext, jose *DAGJOSE, linkSystem ipld.LinkSystem) (ipld.Link, error) {
	return linkSystem.Store(linkContext, LinkPrototype, jose.AsNode())
}

var NodePrototype = new(DAGJOSENodePrototype)

// LoadJOSE is a convenience function which wraps ipld.LinkSystem.Load. This
// will provide the dagjose.NodePrototype to the link system and attempt to
// cast the result to a DAGJOSE object
func LoadJOSE(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (*DAGJOSE, error) {
	n, err := linkSystem.Load(
		linkContext,
		lnk,
		NodePrototype,
	)
	if err != nil {
		return nil, err
	}

	dagJOSE, ok := n.(dagJOSENode)
	if !ok {
		return nil, errors.New("type conversion error")
	}
	return &dagJOSE.DAGJOSE, nil
}
