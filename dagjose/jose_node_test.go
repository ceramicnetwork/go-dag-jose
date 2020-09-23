package dagjose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
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

func sliceOfBytes() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) []byte {
        return rapid.SliceOf(rapid.Byte()).Draw(t, "A slice of bytes").([]byte)
    })
}

func ipldStringGen() *rapid.Generator {
    return  rapid.Custom(func(t* rapid.T) ipld.Node {
        return basicnode.NewString(rapid.String().Draw(t, "an IPLD string").(string))
    })
}

func ipldFloatGen() *rapid.Generator {
    return rapid.Custom(func(t * rapid.T) ipld.Node {
        return basicnode.NewFloat(rapid.Float64().Draw(t, "an IPLD float").(float64))
    })
}

func ipldIntGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        return basicnode.NewInt(rapid.Int().Draw(t, "an IPLD integer").(int))
    })
}

func ipldNullGen() *rapid.Generator {
    return rapid.Just(ipld.Null)
}

func ipldBoolGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        return basicnode.NewBool(rapid.Bool().Draw(t, "an IPLD bool").(bool))
    })
}

func ipldBytesGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        return basicnode.NewBytes(rapid.SliceOf(rapid.Byte()).Draw(t, "some IPLD Bytes").([]byte))
    })
}

func ipldListGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        elems := rapid.SliceOf(ipldNodeGen()).Draw(t, "elements of an IPLD list").([]ipld.Node)
        return fluent.MustBuildList(
            basicnode.Prototype.List,
            len(elems),
            func(la fluent.ListAssembler) {
                for _, elem := range elems {
                    la.AssembleValue().AssignNode(elem)
                }
            },
        )
    })
}

func ipldMapGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        keys := rapid.SliceOf(rapid.String()).Draw(t, "IPLD map keys").([]string)
        return fluent.MustBuildMap(
            basicnode.Prototype.Map,
            len(keys),
            func(ma fluent.MapAssembler) {
                for _, key := range keys {
                    value := ipldNodeGen().Draw(t, "an IPLD map value").(ipld.Node)
                    ma.AssembleEntry(key).AssignNode(value)
                }
            },
        )
    })
}

func ipldNodeGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) ipld.Node {
        return rapid.OneOf(
            ipldStringGen(),
            ipldIntGen(),
            ipldFloatGen(),
            ipldNullGen(),
            ipldBoolGen(),
            ipldListGen(),
            ipldMapGen(),
        ).Draw(t, "an IPLD node").(ipld.Node)
    })
}

func signatureGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) JOSESignature {
        return JOSESignature{
            protected: sliceOfBytes().Draw(t, "signaure protected bytes").([]byte),
            header: ipldMapGen().Draw(t, "signature header").(map[string]ipld.Node),
            signature: sliceOfBytes().Draw(t, "signature bytes").([]byte),
        }
    })
}

func recipientGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) JWERecipient {
        return JWERecipient{
            header: ipldMapGen().Draw(t, "recipient header").(map[string]ipld.Node),
            encrypted_key: sliceOfBytes().Draw(t, "recipient encrypted key").([]byte),
        }
    })
}

func arbitraryJoseGen() *rapid.Generator {
    return rapid.Custom(func(t *rapid.T) DagJOSE {
        return DagJOSE{
            payload: sliceOfBytes().Draw(t, "jose payload").([]byte),
            signatures: rapid.SliceOf(signatureGen()).Draw(t, "jose signatures").([]JOSESignature),
            protected: sliceOfBytes().Draw(t, "jose protected").([]byte),
            unprotected: sliceOfBytes().Draw(t, "jose unprotected").([]byte),
            iv: sliceOfBytes().Draw(t, "JOSE iv").([]byte),
            aad: sliceOfBytes().Draw(t, "JOSE iv").([]byte),
            ciphertext: sliceOfBytes().Draw(t, "JOSE iv").([]byte),
            tag: sliceOfBytes().Draw(t, "JOSE iv").([]byte),
            recipients: rapid.SliceOf(recipientGen()).Draw(t, "JOSE recipients").([]JWERecipient),
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

func TestRoundTripArbitraryJOSE(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        jose := arbitraryJoseGen().Draw(t, "An arbitrary JOSE object").(DagJOSE)
        roundTripped := roundTripJose(&jose)
        require.Equal(t, jose, roundTripped)
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
