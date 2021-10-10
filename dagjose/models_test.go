package dagjose

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

// This test suite is mostly a set of property based tests for the serialization
// of JOSE objects to and from IPLD and JSON. Serialization is well suited to
// property based testing as we have the very straightforward property that
// serializing followed by deserialization should be the same as the identity
// function.
//
// In order to test this property we use the `rapid` property testing library.
// We start by defining a series of generators, which we use to generate arbitrary
// JOSE objects

// Generate an arbitrary CID
func cidGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *cid.Cid {
		data := rapid.SliceOfN(rapid.Byte(), 10, 100).Draw(t, "cid data bytes").([]byte)
		mh, err := multihash.Sum(data, multihash.SHA3_384, 48)
		if err != nil {
			// TODO log err
			return nil
		}
		result := cid.NewCidV1(cid.Raw, mh)
		return &result
	})
}

// An arbitrary ed25519 private key
func ed25519PrivateKeyGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ed25519.PrivateKey {
		seedBytes := rapid.ArrayOf(ed25519.SeedSize, rapid.Byte()).Draw(t, "private key bytes").([ed25519.SeedSize]byte)
		return ed25519.NewKeyFromSeed(seedBytes[:])
	})
}

// Represents a JWS which has been signed and the private key used to sign it
type ValidJWS struct {
	dagJOSE *DAGJWS
	key     ed25519.PrivateKey
}

// Generate a signed JWS along with the private key used to sign it
func validJWSGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *ValidJWS {
		link := cidGen().Draw(t, "Valid DAGJOSE payload").(*cid.Cid)
		privateKey := ed25519PrivateKeyGen().Draw(t, "valid jws private key").(ed25519.PrivateKey)

		signer, err := gojose.NewSigner(gojose.SigningKey{
			Algorithm: gojose.EdDSA,
			Key:       privateKey,
		}, nil)
		if err != nil {
			// "error creating signer for ValidJWS"
			return nil
		}
		goJOSEJWS, err := signer.Sign(link.Bytes())
		if err != nil {
			// "Error signing ValidJWS"
			return nil
		}
		dagJOSE, err := ParseJWS([]byte(goJOSEJWS.FullSerialize()))
		if err != nil {
			// "error creating dagJOSE"
			return nil
		}
		return &ValidJWS{
			dagJOSE: dagJOSE,
			key:     privateKey,
		}
	})
}

// Generate a non-empty slice of JWSSignatures. Note that the signatures are
// not valid, they are just arbitrary byte sequences.
func sliceOfSignatures() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []jwsSignature {
		return rapid.SliceOf(signatureGen()).Filter(func(sigs interface{}) bool {
			return len(sigs.([]jwsSignature)) > 0
		}).Draw(t, "A non empty slice of signatures").([]jwsSignature)
	})
}

// Generate a non empty slice of JWERecipients
func sliceOfRecipients() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []jweRecipient {
		return rapid.SliceOf(recipientGen()).Filter(func(recipients interface{}) bool {
			return len(recipients.([]jweRecipient)) > 0
		}).Draw(t, "A nillable slice of signatures").([]jweRecipient)
	})
}

// Generate a slice of bytes, the slice may be nil
func sliceOfBytes() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []byte {
		isNil := rapid.Bool().Draw(t, "").(bool)
		if isNil {
			return nil
		}
		return rapid.SliceOf(rapid.Byte()).Draw(t, "A nillable slice of bytes").([]byte)
	})
}

// Generate a non nilable slice of bytes
func nonNilSliceOfBytes() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) []byte {
		return rapid.SliceOf(rapid.Byte()).Draw(t, "A slice of bytes").([]byte)
	})
}

// Below are a series of generators for ipld.Node's. This is required because
// the unprotected headers of both JWE and JWS objects can contain arbitrary
// JSON, which we translate into basic node objects

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
		return basicnode.NewInt(rapid.Int64().Draw(t, "an IPLD integer").(int64))
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

// Generate a list of arbitrary ipld.Node, the depth parameter is decreased by one
// and passed to the generator of the child nodes.
func ipldListGen(depth int) *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) ipld.Node {
		elems := rapid.SliceOf(ipldNodeGen(depth-1)).Draw(t, "elements of an IPLD list").([]ipld.Node)
		return fluent.MustBuildList(
			basicnode.Prototype.List,
			int64(len(elems)),
			func(la fluent.ListAssembler) {
				for _, elem := range elems {
					la.AssembleValue().AssignNode(elem)
				}
			},
		)
	})
}

// Generate a map of arbitrary ipld Nodes, the depth parameter is decreased by
// one and passed to the generator of child nodes.
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
			int64(len(keys)),
			func(ma fluent.MapAssembler) {
				for _, key := range keys {
					value := ipldNodeGen(depth-1).Draw(t, "an IPLD map value").(ipld.Node)
					ma.AssembleEntry(key).AssignNode(value)
				}
			},
		)
	})
}

// Generate a map of ipld nodes with string keys. This is used for the top
// level of the unprotected header of JOSE objects.
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

// Generate an arbitrary IPLD node. The depth parameter is used to determine
// the maximum depth of recursive data types (map and list) generated by this
// generator.
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

// Generate an arbitrary JWSSignature, note that the signature is not valid
func signatureGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) jwsSignature {
		return jwsSignature{
			protected: sliceOfBytes().Draw(t, "signature protected bytes").([]byte),
			header:    stringKeyedIPLDMapGen(4).Draw(t, "signature header").(map[string]ipld.Node),
			signature: nonNilSliceOfBytes().Draw(t, "signature bytes").([]byte),
		}
	})
}

// Generate an arbitrary JWERecipient
func recipientGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) jweRecipient {
		return jweRecipient{
			header:       stringKeyedIPLDMapGen(4).Draw(t, "recipient header").(map[string]ipld.Node),
			encryptedKey: sliceOfBytes().Draw(t, "recipient encrypted key").([]byte),
		}
	}).Filter(func(recipient jweRecipient) bool {
		return recipient.encryptedKey != nil || recipient.header != nil
	})
}

// Generate an arbitrary JWS, note that the signatures will not  be valid
func jwsGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *DAGJWS {
		return (DAGJOSE{
			payload:    cidGen().Draw(t, "a JWS CID").(*cid.Cid),
			signatures: sliceOfSignatures().Draw(t, "JOSE signatures").([]jwsSignature),
		}).AsJWS()
	})
}

// Generate an arbitrary JWE, note that the ciphertext is just random bytes and
// cannot be decrypted to anything
func jweGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *DAGJWE {
		return (&DAGJOSE{
			protected:   sliceOfBytes().Draw(t, "JOSE protected").([]byte),
			unprotected: sliceOfBytes().Draw(t, "JOSE unprotected").([]byte),
			iv:          sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			aad:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			ciphertext:  nonNilSliceOfBytes().Draw(t, "JOSE iv").([]byte),
			tag:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			recipients:  sliceOfRecipients().Draw(t, "JOSE recipients").([]jweRecipient),
		}).AsJWE()
	})
}

// Generate an arbitrary JOSE object, i.e either a JWE or a JWS
func arbitraryJOSEGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *DAGJOSE {
		isJWE := rapid.Bool().Draw(t, "whether this JOSE is a JWE").(bool)
		if isJWE {
			return jweGen().Draw(t, "an arbitrary JWE").(*DAGJWE).AsJOSE()
		} else {
			return jwsGen().Draw(t, "an arbitrary JWS").(*DAGJWS).AsJOSE()
		}
	})
}

// Generate a JWS with only one signature
func singleSigJWSGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *DAGJWS {
		return (&DAGJOSE{
			payload: cidGen().Draw(t, "a JWS CID").(*cid.Cid),
			signatures: []jwsSignature{
				signatureGen().Draw(t, "").(jwsSignature),
			},
		}).AsJWS()
	})
}

// Generate a JWE with only one recipient
func singleRecipientJWEGen() *rapid.Generator {
	return rapid.Custom(func(t *rapid.T) *DAGJWE {
		return (&DAGJOSE{
			protected:   sliceOfBytes().Draw(t, "jose protected").([]byte),
			unprotected: sliceOfBytes().Draw(t, "jose unprotected").([]byte),
			iv:          sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			aad:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			ciphertext:  nonNilSliceOfBytes().Draw(t, "JOSE iv").([]byte),
			tag:         sliceOfBytes().Draw(t, "JOSE iv").([]byte),
			recipients:  []jweRecipient{recipientGen().Draw(t, "JWE recipient").(jweRecipient)},
		}).AsJWE()
	})
}

// Normalize json values contained in the unprotected headers of signatures
// and recipients
//
// Unprotected headers can contain arbitrary JSON. There are two things we have
// to normalize for comparison in tests:
// - Integer values will end up as float values after serialization -> deserialization
//   so we convert all integer values to floats
// - Maps don't have a defined order in JSON, so we modify all maps so that
//   they are ordered by key
func normalizeJoseForJSONComparison(d *DAGJOSE) error {
	for _, recipient := range d.recipients {
		for key, value := range recipient.header {
			node, err := normalizeIPLDNode(value)
			if err != nil {
				return err
			}
			recipient.header[key] = node
		}
	}
	for _, sig := range d.signatures {
		for key, value := range sig.header {
			node, err := normalizeIPLDNode(value)
			if err != nil {
				return err
			}
			sig.header[key] = node
		}
	}
	return nil
}

// This is used by the `normalizeJoseForJSONComparison` function to normalize
// the arbitrary IPLD structures in the headers of JWE and JWS objects
func normalizeIPLDNode(n ipld.Node) (ipld.Node, error) {
	switch n.Kind() {
	case ipld.Kind_Int:
		asInt, err := n.AsInt()
		if err != nil {
			return nil, errors.Wrap(err, "normalizeIPLDNode error calling AsInt")
		}
		return basicnode.NewFloat(float64(asInt)), nil
	case ipld.Kind_Map:
		mapIterator := n.MapIterator()
		if mapIterator == nil {
			return nil, errors.New("normalizeIPLDNode nil MapIterator returned from map node")
		}
		// For a map we normalize such that the map keys are in sorted order,
		// this order is maintained by the basicnode.Map implementation
		return fluent.MustBuildMap(
			n.Prototype(),
			0,
			func(ma fluent.MapAssembler) {
				type kv struct {
					key   string
					value ipld.Node
				}
				var kvs []kv
				for !mapIterator.Done() {
					key, val, err := mapIterator.Next()
					if err != nil {
						return
					}
					keyString, err := key.AsString()
					if err != nil {
						return
					}
					node, err := normalizeIPLDNode(val)
					if err != nil {
						return
					}
					kvs = append(kvs, kv{key: keyString, value: node})
				}
				sort.SliceStable(kvs, func(i int, j int) bool {
					return kvs[i].key < kvs[j].key
				})
				for _, kv := range kvs {
					ma.AssembleKey().AssignString(kv.key)
					ma.AssembleValue().AssignNode(kv.value)
				}
			},
		), nil
	case ipld.Kind_List:
		listIterator := n.ListIterator()
		if listIterator == nil {
			return nil, errors.New("convertIntNodesToFloat nil ListIterator returned from list node")
		}
		return fluent.MustBuildList(
			n.Prototype(),
			0,
			func(la fluent.ListAssembler) {
				for !listIterator.Done() {
					_, val, err := listIterator.Next()
					if err != nil {
						return
					}
					node, err := normalizeIPLDNode(val)
					if err != nil {
						return
					}
					la.AssembleValue().AssignNode(node)
				}
			},
		), nil
	default:
		return n, nil
	}
}

// Given a JOSE object we encode it using BuildJOSELink and decode it using LoadJOSE and return the result
func roundTripJose(j *DAGJOSE) (*DAGJOSE, error) {
	var buf bytes.Buffer
	ls := cidlink.DefaultLinkSystem()
	ls.StorageWriteOpener = func(lnkCtx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		return &buf, func(lnk ipld.Link) error { return nil }, nil
	}
	ls.StorageReadOpener = func(lnkCtx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		return bytes.NewReader(buf.Bytes()), nil
	}

	link, err := StoreJOSE(
		ipld.LinkContext{},
		j,
		ls,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error storing DAGJOSE")
	}
	jose, err := LoadJOSE(
		link,
		ipld.LinkContext{},
		ls,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error reading data from datastore")
	}
	return jose, nil
}

// Check that if we encode and decode a valid JWS object then the
// output is equal to the input (up to ipld normalization)
func TestRoundTripValidJWS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		validJWS := validJWSGen().Draw(t, "valid JWS").(*ValidJWS)
		jose, err := roundTripJose(validJWS.dagJOSE.AsJOSE())
		require.NoError(t, err)
		roundTripped := jose.AsJWS()
		require.Equal(t, validJWS.dagJOSE, roundTripped)
	})
}

// Check that if we encode and decode an arbitrary JOSE object then the
// output is equal to the input (up to ipld normalization)
func TestRoundTripArbitraryJOSE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jose := arbitraryJOSEGen().Draw(t, "An arbitrary JOSE object").(*DAGJOSE)
		roundTripped, err := roundTripJose(jose)
		require.NoError(t, err)
		err = normalizeJoseForJSONComparison(jose)
		require.NoError(t, err)
		err = normalizeJoseForJSONComparison(roundTripped)
		require.NoError(t, err)
		require.Equal(t, jose, roundTripped)
	})
}

// Decoding should always return either a JWS or a JWE if the input is valid
func TestAlwaysDeserializesToEitherJWSOrJWE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jose := arbitraryJOSEGen().Draw(t, "An arbitrary JOSE object").(*DAGJOSE)
		roundTripped, err := roundTripJose(jose)
		require.NoError(t, err)
		if roundTripped.AsJWE() == nil {
			require.NotNil(t, roundTripped.AsJWS())
		}
	})
}

// If we parse the JSON serialization of a JWS then the output should equal
// the input
func TestJSONSerializationJWS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dagJWS := jwsGen().Draw(t, "An arbitrary JWS").(*DAGJWS)
		generalSerialization, err := dagJWS.GeneralJSONSerialization()
		assert.NoError(t, err)
		parsedJOSE, err := ParseJWS(generalSerialization)
		assert.NoError(t, err, "error parsing full serialization")
		err = normalizeJoseForJSONComparison(dagJWS.AsJOSE())
		require.NoError(t, err)
		require.EqualValues(t, dagJWS, parsedJOSE)
	})
}

// If we parse the JSON serialization of a JWE then the output should equal
// the input
func TestJSONSerializationJWE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dagJWE := jweGen().Draw(t, "An arbitrary JOSE object").(*DAGJWE)
		generalSerialization, err := dagJWE.GeneralJSONSerialization()
		parsedJOSE, err := ParseJWE(generalSerialization)
		require.NoError(t, err)
		err = normalizeJoseForJSONComparison(dagJWE.AsJOSE())
		require.NoError(t, err)
		require.EqualValues(t, dagJWE, parsedJOSE)
	})
}

// A JWS without a signature is not valid
func TestMissingPayloadErrorParsingJWS(t *testing.T) {
	jsonStr := "{\"signatures\": []}"
	jws, err := ParseJWS([]byte(jsonStr))
	require.Error(t, err)
	require.Empty(t, jws)
}

// A JWE without ciphertext is not valid
func TestMissingCiphertextErrorParsingJWE(t *testing.T) {
	jsonStr := "{\"header\": {}}"
	jwe, err := ParseJWE([]byte(jsonStr))
	require.Error(t, err)
	require.Empty(t, jwe)
}

// If we parse the flattened serialization of a JWS then the input should
// equal the output
func TestFlattenedSerializationJWS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jws := singleSigJWSGen().Draw(t, "a JWS").(*DAGJWS)
		flattenedSerialization, err := jws.FlattenedSerialization()
		require.NoError(t, err, "error creating flattened serialization")
		parsedJOSE, err := ParseJWS(flattenedSerialization)
		require.NoError(t, err, "error parsing flattenedSerialization")
		err = normalizeJoseForJSONComparison(jws.AsJOSE())
		require.NoError(t, err)
		require.Equal(t, jws, parsedJOSE)
	})
}

// Trying to serialize a JWS with more than one signature to a flattened
// serialization should throw an error
func TestFlattenedJWSErrorIfSignatureAndSignaturesDefined(t *testing.T) {
	jsonStr := "{\"signature\": \"\", \"signatures\": [], \"payload\": \"\"}"
	jws, err := ParseJWS([]byte(jsonStr))
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot contain both a 'signature' and a 'signatures'")
	require.Empty(t, jws)
}

// If we parse the flattened serialization of a JWE then the input should
// equal the output
func TestFlattenedSerializationJWE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jwe := singleRecipientJWEGen().Draw(t, "a JWE with one recipient").(*DAGJWE)
		flattenedSerialization, err := jwe.FlattenedSerialization()
		if err != nil {
			t.Errorf("error creating flattened serialization: %v", err)
			return
		}
		parsedJose, err := ParseJWE(flattenedSerialization)
		if err != nil {
			t.Errorf("error parsing flattenedSerialization: %v", err)
			return
		}
		err = normalizeJoseForJSONComparison(jwe.AsJOSE())
		require.NoError(t, err)
		require.Equal(t, jwe, parsedJose)
	})
}

// If the incoming IPLD data contains a payload which is not a valid CID we
// should raise an error
func TestLoadingJWSWithNonCIDPayloadReturnsError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payload := nonNilSliceOfBytes().Filter(func(payloadBytes []byte) bool {
			_, _, err := cid.CidFromBytes(payloadBytes)
			return err != nil
		}).Draw(t, "A slice of bytes which is not a valid CID").([]byte)
		node := fluent.MustBuildMap(
			basicnode.Prototype.Map,
			2,
			func(ma fluent.MapAssembler) {
				ma.AssembleEntry("payload").AssignBytes(payload)
			},
		)
		buf := bytes.Buffer{}
		ls := cidlink.DefaultLinkSystem()
		ls.StorageWriteOpener = func(lnkCtx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
			return &buf, func(lnk ipld.Link) error { return nil }, nil
		}
		ls.StorageReadOpener = func(lnkCtx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
			return bytes.NewReader(buf.Bytes()), nil
		}
		link, err := ls.Store(
			ipld.LinkContext{},
			LinkPrototype,
			node,
		)
		if err != nil {
			t.Errorf("Error creating link to invalid payload node: %v", err)
			return
		}
		_, err = LoadJOSE(
			link,
			ipld.LinkContext{},
			ls,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "payload is not a valid CID")
	})
}

// Trying to serialize a JWE with more than one recipient to a flattened
// serialization should throw an error
func TestFlattenedJWEErrorIfEncryptedKeyOrHeaderAndRecipientsDefined(t *testing.T) {
	scenarios := [][]byte{
		[]byte("{\"ciphertext\": \"\", \"encrypted_key\": \"\", \"recipients\": []}"),
		[]byte("{\"ciphertext\": \"\", \"header\": {}, \"recipients\": []}"),
	}
	for _, scenario := range scenarios {
		jwe, err := ParseJWE(scenario)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot contain 'recipients' and either 'encrypted_key' or 'header'")
		require.Empty(t, jwe)
	}
}
