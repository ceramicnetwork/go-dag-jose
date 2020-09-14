# dagjose-go

An implementation of the [`dag-jose`](https://github.com/ipld/specs/pull/269) multiformat for Go. 

## Example usage

To read a JWS from IPFS:

```go
import (
    "github.com/alexjg/dagjose" 
    cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)
// Here we're creating a `CID` which points to a JWS
jwsCid, err := cid.Decode(
    "bagcqcerafzecmo67npzj56gfyukh3tbwxywgfsmxirr4oq57s6sddou4p5dq"
)
if err != nil {
    panic(err)
}
// cidlink.Link is an implementation of `ipld.Link` backed by a CID
jwsLnk := cidlink.Link{Cid: jwsCid}


// This is the `NodeBuilder` which knows how to build a `dagjose.DagJOSE` object
builder := dagjose.NewBuilder()
err = lnk.Load(
    context.Background(),
    ipld.LinkContext{},
    builder,
    <an implementation of ipld.Loader, which knows how to get the block data from IPFS>,
)
if err != nil {
    panic(err)
}
n := builder.Build()
jwsNode := n.(*dagjose.DagJOSE), nil
```

To write a JWS to IPFS

```go
dagJws, err := dagjose.NewDagJWS("<the general JSON serialization of a JWS>")
if err != nil {
    panic(err)
}
linkBuilder := cidlink.LinkBuilder{Prefix: cid.Prefix{
    Version:  1,    // Usually '1'.
    Codec:    0x85, // 0x71 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
    MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
    MhLength: 48,   // sha3-224 hash has a 48-byte sum.
}}
link, err := linkBuilder.Build(
    context.Background(),
    ipld.LinkContext{},
    dagJws,
    <an implementation of `ipld.Storer` which knows how to store the raw block data>
)
if err != nil {
    panic(err)
}
```
