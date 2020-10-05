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

jose, err := dagjose.LoadJOSE(
    jwsLnk,
    context.Background(),
    ipld.LinkContext{},
    <an implementation of ipld.Loader, which knows how to get the block data from IPFS>,
)
if err != nil {
    panic(err)
}
if jose.AsJWS() != nil {
    // We have a JWS object, print the general serialization of it
    print(jose.AsJWS().GeneralJSONSerialization())
}
```

To write a JWS to IPFS

```go
dagJws, err := dagjose.ParseJWS("<the general JSON serialization of a JWS>")
if err != nil {
    panic(err)
}
err = dagjose.BuildJOSELink(
    context.Background(),
    ipld.LinkContext{},
    dagJws,
    <an implementation of `ipld.Storer` which knows how to store the raw block data>
)
if err != nil {
    panic(err)
}
```
