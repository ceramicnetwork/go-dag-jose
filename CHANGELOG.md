### v0.0.5

Update to `go-ipld-prime` 0.9.0. `go-ipld-prime` now uses a `LinkSystem`
abstraction which combines the `ipld.Storer` and `ipld.Loader` interfaces.
Consequently the convenience functions `dagjose.BuildJOSELink` and 
`dagjose.LoadJOSE` must change to use this new interface. We also expose a
`dagjose.LinkPrototype` which knows how to build CID links to a dag-jose object
for users who are using the underlying `LinkSystem` directly.

To change existing code to be compatible with the new API note the following:

1. `dagjose.BuildJOSELink` has become `StoreJOSE`. This accepts a `LinkSystem` in place
   of the previous `ipld.Storer`
2. `dagjose.LoadJOSE` accepts a `LinkSystem` in place of the previous `ipld.Loader`

To call these you will now need to create a `LinkSystem`, typically this will
look like this:

```go
import (
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

ls := cidlink.DefaultLinkSystem()
ls.StorageWriteOpener = <whatever you were previously using as ipld.Storer>
ls.StorageReadOpener = <whatever you were previously using as ipld.Loader>

link, err := dagjose.StoreJOSE(
    ipld.LinkContext{},
    j,
    ls,
)

jose, err := dagjose.LoadJOSE(
    link,
    ipld.LinkContext{},
    ls,
)
```
