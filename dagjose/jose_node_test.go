package dagjose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
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

func sliceOfSignatures() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []JWSSignature {
		isNil := rapid.Bool().Draw(t, "").(bool)
		if isNil {
			return nil
		}
		return rapid.SliceOf(signatureGen()).Draw(t, "A nillable slice of bytes").([]JWSSignature)
	})
}

func sliceOfBytes() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []byte {
		isNil := rapid.Bool().Draw(t, "").(bool)
		if isNil {
			return nil
		}
		return rapid.SliceOf(rapid.Byte()).Draw(t, "A nillable slice of bytes").([]byte)
	})
}

func nonNilSliceOfBytes() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []byte {
		return rapid.SliceOf(rapid.Byte()).Draw(t, "A slice of bytes").([]byte)
	})
}

func ipldStringGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
		return basicnode.NewString(rapid.String().Draw(t, "an IPLD string").(string))
	})
}

func ipldFloatGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
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

func ipldListGen(depth int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
		elems := rapid.SliceOf(ipldNodeGen(depth-1)).Draw(t, "elements of an IPLD list").([]ipld.Node)
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

func ipldMapGen(depth int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
		keys := rapid.SliceOfDistinct(
			rapid.String(),
			func(k string) string {
				return k
			},
		).Draw(t, "IPLD map keys").([]string)
		return fluent.MustBuildMap(
			basicnode.Prototype.Map,
			len(keys),
			func(ma fluent.MapAssembler) {
				for _, key := range keys {
					value := ipldNodeGen(depth-1).Draw(t, "an IPLD map value").(ipld.Node)
					ma.AssembleEntry(key).AssignNode(value)
				}
			},
		)
	})
}

func stringKeyedIPLDMapGen(depth int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) map[string]ipld.Node {
		isNil := rapid.Bool().Draw(t, "whether the map is nil").(bool)
		if isNil {
			return nil
		}
		keys := rapid.SliceOf(rapid.String()).Draw(t, "IPLD map keys").([]string)
		result := make(map[string]ipld.Node)
		for _, key := range keys {
			value := ipldNodeGen(depth-1).Draw(t, "an IPLD map value").(ipld.Node)
			result[key] = value
		}
		return result
	})
}

func ipldNodeGen(depth int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
		elems := []*rapid.Generator{
			ipldStringGen(),
			ipldIntGen(),
			ipldFloatGen(),
			ipldNullGen(),
			ipldBoolGen(),
		}
		if depth > 0 {
			elems = append(elems, ipldListGen(depth), ipldMapGen(depth))
		}
		return rapid.OneOf(
			elems...,
		).Draw(t, "an IPLD node").(ipld.Node)
	})
}

func signatureGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) JWSSignature {
		return JWSSignature{
			protected: sliceOfBytes().Draw(t, "signature protected bytes").([]byte),
			header:    stringKeyedIPLDMapGen(4).Draw(t, "signature header").(map[string]ipld.Node),
			signature: nonNilSliceOfBytes().Draw(t, "signature bytes").([]byte),
		}
	})
}

func recipientGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) JWERecipient {
		return JWERecipient{
			header:        stringKeyedIPLDMapGen(4).Draw(t, "recipient header").(map[string]ipld.Node),
			encrypted_key: sliceOfBytes().Draw(t, "recipient encrypted key").([]byte),
		}
	})
}

func arbitraryJoseGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) DagJOSE {
		return DagJOSE{
			payload:     sliceOfBytes().Draw(t, "jose payload").([]byte),
			signatures:  sliceOfSignatures().Draw(t, "jose signatures").([]JWSSignature),
			protected:   sliceOfBytes().Draw(t, "jose protected").([]byte),
			unprotected: sliceOfBytes().Draw(t, "jose unprotected").([]byte),
			iv:          sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			aad:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			ciphertext:  sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			tag:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			recipients:  rapid.SliceOf(recipientGen()).Draw(t, "JOSE recipients").([]JWERecipient),
		}
	})
}

func singleSigJWSGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) DagJOSE {
		return DagJOSE{
			payload: sliceOfBytes().Draw(t, "jose payload").([]byte),
			signatures: []JWSSignature{
				signatureGen().Draw(t, "").(JWSSignature),
			},
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
		require.Equal(t, jose, *roundTripped)
	})
}

func TestGeneralJSONSerialization(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jose := arbitraryJoseGen().Draw(t, "An arbitrary JOSE object").(DagJOSE)
		generalSerialization := jose.GeneralJSONSerialization()
		parsedJose, err := NewDagJWS(generalSerialization)
		normalizeJoseForJsonComparison(&jose)
		if err != nil {
			t.Errorf("error parsing full serialization: %v", err)
		}
		require.Equal(t, &jose, parsedJose)
	})
}

//func TestFlattenedSerialization(t *testing.T) {
//rapid.Check(t, func(t *rapid.T) {
//jose := arbitraryJoseGen().Draw(t, "").(DagJOSE)
//flattenedSerialization, err := jose.FlattenedSerialization()
//if err != nil {
//t.Errorf("error creating flattened serialization: %v", err)
//}
//parsedJose, err := NewDagJWS(flattenedSerialization)
//if err != nil {
//t.Errorf("error parsing flattenedSerialization: %v", err)
//}
//require.Equal(t, &jose, parsedJose)
//})
//}

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

// Normalize json values contained in the unprotected headers of signatures
// and recipients
//
// Unprotected headers can contain arbitrary JSON. There are two things we have
// to normalise for comparison:
// - Integer values will end up as float values after serialization -> deserialization
//   so we convert all integer values to floats
// - Maps don't have a defined order in JSON, so we modify all maps so that
//   they are ordered by key
func normalizeJoseForJsonComparison(d *DagJOSE) {
	for _, recipient := range d.recipients {
		for key, value := range recipient.header {
			recipient.header[key] = normalizeIpldNode(value)
		}
	}
	for _, sig := range d.signatures {
		for key, value := range sig.header {
			sig.header[key] = normalizeIpldNode(value)
		}
	}
}

func normalizeIpldNode(n ipld.Node) ipld.Node {
	switch n.ReprKind() {
	case ipld.ReprKind_Int:
		asInt, err := n.AsInt()
		if err != nil {
			panic(fmt.Errorf("normalizeIpldNode error calling AsInt: %v", err))
		}
		return basicnode.NewFloat(float64(asInt))
	case ipld.ReprKind_Map:
		mapIterator := n.MapIterator()
		if mapIterator == nil {
			panic(fmt.Errorf("normalizeIpldNode nil MapIterator returned from map node"))
		}
		return fluent.MustBuildMap(
			basicnode.Prototype.Map,
			0,
			func(ma fluent.MapAssembler) {
				type kv struct {
					key   string
					value ipld.Node
				}
				kvs := make([]kv, 0)
				for !mapIterator.Done() {
					key, val, err := mapIterator.Next()
					if err != nil {
						panic(fmt.Errorf("normalizeIpldNode error calling Next on mapiterator: %v", err))
					}
					keyString, err := key.AsString()
					if err != nil {
						panic(fmt.Errorf("normalizeIpldNode: error converting key to string: %v", err))
					}
					kvs = append(kvs, kv{key: keyString, value: normalizeIpldNode(val)})
				}
				sort.SliceStable(kvs, func(i int, j int) bool {
					return kvs[i].key < kvs[j].key
				})
				for _, kv := range kvs {
					ma.AssembleKey().AssignString(kv.key)
					ma.AssembleValue().AssignNode(kv.value)
				}
			},
		)
	case ipld.ReprKind_List:
		listIterator := n.ListIterator()
		if listIterator == nil {
			panic(fmt.Errorf("convertIntNodesToFlaot nil ListIterator returned from list node"))
		}
		return fluent.MustBuildList(
			basicnode.Prototype.List,
			0,
			func(la fluent.ListAssembler) {
				for !listIterator.Done() {
					_, val, err := listIterator.Next()
					if err != nil {
						panic(fmt.Errorf("convertIntNodesToFlaot error calling Next on listiterator: %v", err))
					}
					la.AssembleValue().AssignNode(normalizeIpldNode(val))
				}
			},
		)
	default:
		return n
	}
}
