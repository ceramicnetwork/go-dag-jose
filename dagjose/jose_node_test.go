package dagjose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
	gojose "gopkg.in/square/go-jose.v2"
	"pgregory.net/rapid"
)

func cidGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) cid.Cid {
		data := rapid.SliceOfN(rapid.Byte(), 10, 100).Draw(t, "cid data bytes").([]byte)
		mh, err := multihash.Sum(data, multihash.SHA3_384, 48)
		if err != nil {
			panic(err)
		}
		return cid.NewCidV1(cid.Raw, mh)
	})
}

func ed25518PrivateKeyGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ed25519.PrivateKey {
		seedBytes := rapid.ArrayOf(ed25519.SeedSize, rapid.Byte()).Draw(t, "private key bytes").([ed25519.SeedSize]byte)
        return ed25519.NewKeyFromSeed(seedBytes[:])
	})
}

type ValidJWS struct {
	dagJose *DagJOSE
	key     ed25519.PrivateKey
}

func validJWSGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ValidJWS {
		link := cidGen().Draw(t, "Valid DagJOSE payload").(cid.Cid)
		privateKey := ed25518PrivateKeyGen().Draw(t, "valid jws private key").(ed25519.PrivateKey)

		signer, err := gojose.NewSigner(gojose.SigningKey{
			Algorithm: gojose.EdDSA,
			Key:       privateKey,
		}, nil)
		if err != nil {
			panic(fmt.Errorf("error creating signer for ValidJWS: %v", err))
		}
		gojoseJws, err := signer.Sign(link.Bytes())
		if err != nil {
			panic(fmt.Errorf("Error signing ValidJWS: %v", err))
		}
		dagJose, err := NewDagJWS(gojoseJws.FullSerialize())
		if err != nil {
			panic(fmt.Errorf("error creating dagjose: %v", err))
		}
		return ValidJWS{
			dagJose: dagJose,
			key:     privateKey,
		}
	})
}

func TestRoundTripValidJWS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		validJws := validJWSGen().Draw(t, "valid JWS").(ValidJWS)
		roundTripped := roundTripJose(validJws.dagJose)
		require.Equal(t, validJws.dagJose, roundTripped)
	})
}

func roundTripJose(j *DagJOSE) *DagJOSE {
    buf := bytes.Buffer{}
    linkBuilder := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,
		Codec:    0x85,
		MhType:   multihash.SHA3_384,
		MhLength: 48,
	}}
    link, err := linkBuilder.Build(
        context.Background(),
        ipld.LinkContext{},
        j,
        func(ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
            return &buf, func(l ipld.Link) error { return nil }, nil
        },
    )
    if err != nil {
        panic(fmt.Errorf("error storing DagJOSE: %v", err))
    }
    nodeBuilder := NewBuilder()
	err = link.Load(
		context.Background(),
		ipld.LinkContext{},
		nodeBuilder,
        func(l ipld.Link, _ ipld.LinkContext) (io.Reader, error) {
            return bytes.NewBuffer(buf.Bytes()), nil
        },
	)
	if err != nil {
        panic(fmt.Errorf("error reading data from datastore: %v", err))
	}
	n := nodeBuilder.Build()
	return n.(*DagJOSE)
}
