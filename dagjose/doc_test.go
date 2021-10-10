package dagjose_test

import (
	"fmt"

	"github.com/alexjg/go-dag-jose/dagjose"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func Example_read() {
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
}

func Example_write() {
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
}
