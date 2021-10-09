package go_dag_jose_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/alexjg/go-dag-jose/dagjose"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func TestRead(t *testing.T) {
	// Here we're creating a `CID` which points to a JWS
	jwsCID, err := cid.Decode("some cid")
	assert.NoError(t, err)

	// cidlink.Link is an implementation of `ipld.Link` backed by a CID
	jwsLnk := cidlink.Link{Cid: jwsCID}

	ls := cidlink.DefaultLinkSystem()

	jose, err := dagjose.LoadJOSE(
		jwsLnk,
		ipld.LinkContext{},
		ls, //<an implementation of ipld.Loader, which knows how to get the block data from IPFS>,
	)
	assert.NoError(t, err)

	asJWS := jose.AsJWS()
	assert.NotEmpty(t, asJWS)

	// We have a JWS object, print the general serialization of it
	fmt.Print(jose.AsJWS().GeneralJSONSerialization())
}

func TestWrite(t *testing.T) {
	dagJws, err := dagjose.ParseJWS([]byte("<the general JSON serialization of a JWS>"))
	assert.NoError(t, err)

	ls := cidlink.DefaultLinkSystem()
	link, err := dagjose.StoreJOSE(
		ipld.LinkContext{},
		dagJws.AsJOSE(),
		ls,
	)
	assert.NoError(t, err)

	fmt.Printf("Link is: %v", link)
}
