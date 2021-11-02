package dagjose

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	//"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	//"github.com/ipld/go-ipld-prime/schema"
)

// A link prototype which will build CIDs using the dag-jose multicodec and
// the sha-384 multihash
var LinkPrototype = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x85, // 0x85 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhLength: 48,   // sha3-224 hash has a 48-byte sum.
}}

// A convenience function which passes the correct dagjose.LinkProtoype and
// DAG-JOSE object to ipld.LinkSystem.Store
func StoreJOSE(linkContext ipld.LinkContext, jose _JOSE, linkSystem ipld.LinkSystem) (ipld.Link, error) {
	return linkSystem.Store(linkContext, LinkPrototype, &jose)
}

// LoadJOSE is a convenience function which wraps ipld.LinkSystem.Load. This
// will provide the dagjose.NodePrototype to the link system and attempt to
// cast the result to a DagJOSE object
func LoadJOSE(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (JOSE, error) {
	n, err := linkSystem.Load(
		linkContext,
		lnk,
		Type.JOSE,
	)
	if err != nil {
		return nil, err
	}

	return n.(JOSE), nil
}
