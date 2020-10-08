package dagjose

import (
	"context"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

// This is a union of the DagJWE and DagJWS types. Typically you will want to
// use AsJWE and AsJWS to get a concrete JOSE object.
type DagJOSE struct {
	// JWS top level keys
	payload    *cid.Cid
	signatures []JWSSignature
	// JWE top level keys
	protected   []byte
	unprotected []byte
	iv          []byte
	aad         []byte
	ciphertext  []byte
	tag         []byte
	recipients  []JWERecipient
}

type JWSSignature struct {
	protected []byte
	header    map[string]ipld.Node
	signature []byte
}

type JWERecipient struct {
	header        map[string]ipld.Node
	encrypted_key []byte
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

// This exposes a similar interface to the cidlink.LinkBuilder from go-ipld-prime. It's primarily a convenience
// function so you don't have to specify the codec version yourself
func BuildJOSELink(ctx context.Context, linkContext ipld.LinkContext, jose *DagJOSE, storer ipld.Storer) (ipld.Link, error) {
	lb := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x85, // 0x71 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 48,   // sha3-224 hash has a 48-byte sum.
	}}
	return lb.Build(
		ctx,
		linkContext,
		jose,
		storer,
	)
}

// LoadJOSE is a convenience function which wraps ipld.Link.Load. This will provide the dagjose.NodeBuilder
// to the link and attempt to cast the result to a DagJOSE object
func LoadJOSE(lnk ipld.Link, ctx context.Context, linkContext ipld.LinkContext, loader ipld.Loader) (*DagJOSE, error) {
	builder := NewBuilder()
	err := lnk.Load(
		ctx,
		linkContext,
		builder,
		loader,
	)
	if err != nil {
		return nil, err
	}

	n := builder.Build()
	return n.(*DagJOSE), nil
}
