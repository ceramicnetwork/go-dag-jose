# go-dag-jose

An implementation of the [`dag-jose`](https://github.com/ipld/specs/pull/269) multiformat for Go. 

## Example usage

To read a JWS from IPFS:

```go 
import ( 
    "github.com/ceramicnetwork/dagjose"
    cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)
// Here we're creating a `CID` which points to a JWS
jwsCid, err := cid.Decode("some cid")
if err != nil {
    panic(err)
}
// cidlink.Link is an implementation of `ipld.Link` backed by a CID
jwsLnk := cidlink.Link{Cid: jwsCid}

ls := cidlink.DefaultLinkSystem()

jose, err := dagjose.LoadJOSE(
    jwsLnk,
    ipld.LinkContext{},
    ls, //<an implementation of ipld.Loader, which knows how to get the block data from IPFS>,
)
if err != nil {
    panic(err)
}
if jose.AsJWS() != nil {
    // We have a JWS object, print the general serialization of it
    print(jose.AsJWS().GeneralJSONSerialization())
} else {
    print("This is not a JWS")
}
```

To write a JWS to IPFS

```go
dagJws, err := dagjose.ParseJWS([]byte("<the general JSON serialization of a JWS>"))
if err != nil {
    panic(err)
}
ls := cidlink.DefaultLinkSystem()
link, err := dagjose.StoreJOSE(
    ipld.LinkContext{},
    dagJws.AsJOSE(),
    ls,
)
if err != nil {
    panic(err)
}
fmt.Printf("Link is: %v", link)
```

## Changelog

This project attempts to stay up to date with changes in `go-ipld-prime`, this
means somewhat frequent API breakage as `go-ipld-prime` is not yet stable. 
See [the changelog](./CHANGELOG.md).
