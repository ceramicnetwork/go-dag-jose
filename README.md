# go-dag-jose

An implementation of the [`dag-jose`](https://github.com/ipld/specs/pull/269) multiformat for Go. 

# Usage

Data that is encoded using the `dag-jose` multicodec is guaranteed to be a CBOR encoding of the general serialization 
of either a JWE or a JWS. In order to  access this information using `go-ipld-prime` we need two things:
* We need to register an encoder and a decoder which will be used by the `cidlink` package to decode the raw data into the IPLD data model
* An implementation of `ipld.NodeAssembler` which knows how to interpret the IPLD data into some concrete go data type which implements `ipld.Node`

The first of these points is handled by importing this package. There is a side  effecting operation in the module
initialization which registers the encoder and decoder with go-ipld-prime.

The latter point is provided by the `dagjose.DagJOSE` data type. This type represents the union of the `dagjose.DAGJWS`
and `dagjose.DAGJWE` types. Typically, you will use `dagjose.LoadJOSE(..)` to load a `dagjose.DagJOSE` object, then 
you will use `DAGJOSE.AsJWS` and `DAGJOSE.AsJWE` to determine whether you have a JWS or JWE object respectively.

This package does not provide any direct access to the fields of the JWS and JWE objects, instead each kind of JOSE 
object has a `GeneralJSONSerialization` method which can be used to obtain the general json serialization to be
passed to third party libraries.

## Example usage

To read a JWS from IPFS:

```go 
import ( 
    "github.com/alexjg/dagjose" cidlink "github.com/ipld/go-ipld-prime/linking/cid"
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
