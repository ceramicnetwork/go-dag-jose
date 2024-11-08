package dagjose

import (
	"bytes"
	"errors"
	"fmt"
	gojose "github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
	"io"
	"pgregory.net/rapid"
	"reflect"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multihash"
)

// This test suite is mostly a set of property based tests for the serialization of JOSE objects to and from IPLD and
// JSON. Serialization is well suited to property based testing as we have the very straightforward property that
// serializing followed by deserialization should be the same as the identity function.
//
// In order to test this property we use the `rapid` property testing library. We start by defining a series of
// generators, used to generate arbitrary JOSE objects.

// A link prototype which will build CIDs using the dag-jose multicodec and the sha-384 multihash
var dagJOSELink = cidlink.LinkPrototype{Prefix: cid.Prefix{
	Version:  1,    // Usually '1'.
	Codec:    0x85, // 0x85 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhType:   0x12, // 0x12 means "sha2-256" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	MhLength: 32,   // sha2-256 hash has a 32-byte sum.
}}

// storeJOSE is a convenience function that passes the correct DAG-JOSE link prototype and DAG-JOSE object to
// ipld.LinkSystem.Store
func storeJOSE(linkContext ipld.LinkContext, jose datamodel.Node, linkSystem ipld.LinkSystem) (ipld.Link, error) {
	return linkSystem.Store(linkContext, dagJOSELink, jose)
}

// loadJOSE is a convenience function that provides the DAG-JOSE node prototype to ipld.LinkSystem.Load and attempts to
// cast the result to a DAG-JOSE object.
func loadJOSE(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (_ datamodel.Node, err error) {
	n, err := loadJWE(lnk, linkContext, linkSystem)
	if err != nil {
		// If there was an error during JWE decode, try decoding as JWS
		n, err = loadJWS(lnk, linkContext, linkSystem)
		if err != nil {
			return nil, err
		}
	}
	return n.(schema.TypedNode).Representation(), nil
}

func loadJWE(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (_ datamodel.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				// A reasonable fallback, for e.g. strings.
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	return linkSystem.Load(
		linkContext,
		lnk,
		Type.DecodedJWE__Repr,
	)
}

func loadJWS(lnk ipld.Link, linkContext ipld.LinkContext, linkSystem ipld.LinkSystem) (_ datamodel.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				// A reasonable fallback, for e.g. strings.
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	return linkSystem.Load(
		linkContext,
		lnk,
		Type.DecodedJWS__Repr,
	)
}

// parseJOSE will return a general form JWE/JWS node given a JSON string representing a JWE/JWS in flattened or general
// serialization
func parseJOSE(jsonBytes []byte) (datamodel.Node, error) {
	buf := bytes.NewReader(jsonBytes)
	anyBuilder := basicnode.Prototype.Any.NewBuilder()
	if err := (dagjson.DecodeOptions{
		ParseLinks: false,
		ParseBytes: false,
	}.Decode(anyBuilder, buf)); err != nil {
		return nil, err
	} else {
		anyNode := anyBuilder.Build()
		if jwe, err := isJWE(anyNode); err != nil {
			return nil, err
		} else if jwe {
			return unflattenJWE(anyNode)
		} else if jws, err := isJWS(anyNode); err != nil {
			return nil, err
		} else if jws {
			return unflattenJWS(anyNode)
		} else {
			return nil, errors.New("invalid JOSE object")
		}
	}
}

// Generate an arbitrary CID
func cidGen() *rapid.Generator[cid.Cid] {
	return rapid.Custom(func(t *rapid.T) cid.Cid {
		return createCid(rapid.SliceOfN(rapid.Byte(), 10, 100).Draw(t, "cid data bytes"))
	})
}

func createCid(data []byte) cid.Cid {
	mh, err := multihash.Sum(data, multihash.SHA3_384, 48)
	if err != nil {
		panic(err)
	}
	return cid.NewCidV1(cid.Raw, mh)
}

// An arbitrary ed25519 private key
func ed25519PrivateKeyGen() *rapid.Generator[ed25519.PrivateKey] {
	return rapid.Custom(func(t *rapid.T) ed25519.PrivateKey {
		seedBytes := rapid.SliceOfN(rapid.Byte(), ed25519.SeedSize, ed25519.SeedSize).Draw(t, "private key bytes")
		return ed25519.NewKeyFromSeed(seedBytes[:])
	})
}

// Generate a signed JWS along with the private key used to sign it
func validJWSGen() *rapid.Generator[datamodel.Node] {
	return rapid.Custom(func(t *rapid.T) datamodel.Node {
		link := cidGen().Draw(t, "Valid DagJOSE payload")
		privateKey := ed25519PrivateKeyGen().Draw(t, "valid jws private key")
		if signer, err := gojose.NewSigner(gojose.SigningKey{
			Algorithm: gojose.EdDSA,
			Key:       privateKey,
		}, nil); err != nil {
			panic(fmt.Errorf("error creating signer for JWS: %v", err))
		} else {
			if joseJws, err := signer.Sign(link.Bytes()); err != nil {
				panic(fmt.Errorf("error signing JWS: %v", err))
			} else if joseNode, err := parseJOSE([]byte(joseJws.FullSerialize())); err != nil {
				panic(fmt.Errorf("error creating dagjose: %v", err))
			} else {
				return joseNode
			}
		}
	})
}

// Generate a non-empty slice of JWSSignatures. Note that the signatures are not valid, they are just arbitrary byte
// sequences.
func signatures(numSignatures int) *rapid.Generator[_EncodedSignatures__Maybe] {
	return rapid.Custom(func(t *rapid.T) _EncodedSignatures__Maybe {
		return _EncodedSignatures__Maybe{
			schema.Maybe_Value,
			_EncodedSignatures{
				rapid.SliceOfN(
					signatureGen(),
					1,
					numSignatures,
				).Draw(t, "A non-empty slice of signatures"),
			},
		}
	})
}

// Generate a slice of JWERecipients
func recipients(numRecipients int) *rapid.Generator[_EncodedRecipients__Maybe] {
	return rapid.Custom(func(t *rapid.T) _EncodedRecipients__Maybe {
		isNil := rapid.Bool().Draw(t, "")
		if isNil {
			return _EncodedRecipients__Maybe{schema.Maybe_Absent, _EncodedRecipients{}}
		}
		return _EncodedRecipients__Maybe{
			schema.Maybe_Value,
			_EncodedRecipients{
				rapid.SliceOfN(
					recipientGen(),
					numRecipients,
					numRecipients,
				).Draw(t, "A non-empty slice of recipients"),
			},
		}
	})
}

// Generate a non-nillable slice of bytes
func nonNilSliceOfBytes() *rapid.Generator[[]byte] {
	return rapid.Custom(func(t *rapid.T) []byte {
		return rapid.SliceOfN(rapid.Byte(), 1, -1).Draw(t, "a slice of bytes")
	})
}

// Generate a slice of bytes, the slice may be nil
func sliceOfRawBytes() *rapid.Generator[_Raw__Maybe] {
	return rapid.Custom(func(t *rapid.T) _Raw__Maybe {
		isNil := rapid.Bool().Draw(t, "")
		if isNil {
			return _Raw__Maybe{schema.Maybe_Absent, _Raw{}}
		}
		return _Raw__Maybe{schema.Maybe_Value, nonNilSliceOfRawBytes().Draw(t, "")}
	})
}

// Generate a non-nillable slice of bytes
func nonNilSliceOfRawBytes() *rapid.Generator[_Raw] {
	return rapid.Custom(func(t *rapid.T) _Raw {
		return _Raw{nonNilSliceOfBytes().Draw(t, "")}
	})
}

// Generate a map of ipld nodes with string keys. This is used for the top
// level of the unprotected header of JOSE objects.
func mapGen() *rapid.Generator[_Any__Maybe] {
	return rapid.Custom(func(t *rapid.T) _Any__Maybe {
		keys := rapid.SliceOfDistinct(
			rapid.StringN(1, -1, -1),
			func(k string) string {
				return k
			},
		).Draw(t, "map keys")
		header := make(map[_String]Any)
		entries := make([]_Map__entry, 0, len(keys))
		for _, key := range keys {
			k := _String{key}
			v := _Any{&_String{string(nonNilSliceOfBytes().Draw(t, "string"))}}
			header[k] = &v
			entries = append(entries, _Map__entry{k, v})
		}
		return _Any__Maybe{schema.Maybe_Value, &_Any{&_Map{header, entries}}}
	})
}

// Generate an arbitrary JWSSignature, note that the signature is not valid
func signatureGen() *rapid.Generator[_EncodedSignature] {
	return rapid.Custom(func(t *rapid.T) _EncodedSignature {
		return _EncodedSignature{
			header:    mapGen().Draw(t, "signature header"),
			protected: sliceOfRawBytes().Draw(t, "signature protected bytes"),
			signature: nonNilSliceOfRawBytes().Draw(t, "signature bytes"),
		}
	})
}

// Generate an arbitrary JWERecipient
func recipientGen() *rapid.Generator[_EncodedRecipient] {
	return rapid.Custom(func(t *rapid.T) _EncodedRecipient {
		return _EncodedRecipient{
			header:        mapGen().Draw(t, "recipient header"),
			encrypted_key: sliceOfRawBytes().Draw(t, "recipient encrypted key"),
		}
	})
}

// Generate an arbitrary JWE, note that the ciphertext is just random bytes and
// cannot be decrypted to anything
func jweGen(numRecipients int) *rapid.Generator[datamodel.Node] {
	return rapid.Custom(func(t *rapid.T) datamodel.Node {
		aad := sliceOfRawBytes().Draw(t, "aad")
		ciphertext := nonNilSliceOfRawBytes().Draw(t, "ciphertext")
		iv := sliceOfRawBytes().Draw(t, "iv")
		protected := sliceOfRawBytes().Draw(t, "protected")
		unprotected := mapGen().Draw(t, "unprotected")
		tag := sliceOfRawBytes().Draw(t, "JOSE iv")
		return &_EncodedJWE__Repr{
			aad,
			ciphertext,
			iv,
			protected,
			recipients(numRecipients).Draw(t, "JWE recipients"),
			tag,
			unprotected,
		}
	})
}

// Generate an arbitrary JWS, note that the signatures will not  be valid
func jwsGen(numSignatures int) *rapid.Generator[datamodel.Node] {
	return rapid.Custom(func(t *rapid.T) datamodel.Node {
		return &_EncodedJWS__Repr{
			payload:    _Raw{cidGen().Draw(t, "a JWS CID").Bytes()},
			signatures: signatures(numSignatures).Draw(t, "JWS signatures"),
		}
	})
}

// Generate an arbitrary JOSE object, i.e. either a JWE or a JWS
func arbitraryJoseGen() *rapid.Generator[datamodel.Node] {
	return rapid.Custom(func(t *rapid.T) datamodel.Node {
		isJwe := rapid.Bool().Draw(t, "whether this jose is a jwe")
		if isJwe {
			return jweGen(-1).Draw(t, "an arbitrary JWE")
		} else {
			return jwsGen(-1).Draw(t, "an arbitrary JWS")
		}
	})
}

// Given a JOSE object we encode it using StoreJOSE and decode it using LoadJOSE and return the result
func roundTripJose(storeJose datamodel.Node) (datamodel.Node, error) {
	buf := bytes.Buffer{}
	ls := cidlink.DefaultLinkSystem()
	ls.StorageWriteOpener = func(lnkCtx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		return &buf, func(lnk ipld.Link) error { return nil }, nil
	}
	ls.StorageReadOpener = func(lnkCtx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		return bytes.NewReader(buf.Bytes()), nil
	}

	// We're going from JOSE -> JOSE, so we don't need to complicate things with `link`.
	multicodec.RegisterDecoder(0x85, DecodeOptions{AddLink: false}.Decode)
	multicodec.RegisterEncoder(0x85, Encode)

	if link, err := storeJOSE(
		ipld.LinkContext{},
		storeJose,
		ls,
	); err != nil {
		panic(fmt.Errorf("error storing DagJOSE: %v", err))
	} else {
		if loadJose, err := loadJOSE(
			link,
			ipld.LinkContext{},
			ls,
		); err != nil {
			panic(fmt.Errorf("error reading data from datastore: %v", err))
		} else {
			return loadJose, nil
		}
	}
}

func compareJOSE(t *rapid.T, encoded datamodel.Node, decoded datamodel.Node) {
	compareJOSEField(t, "aad", encoded, decoded)
	compareJOSEField(t, "ciphertext", encoded, decoded)
	compareJOSEField(t, "iv", encoded, decoded)
	compareJOSEField(t, "payload", encoded, decoded)
	compareJOSEField(t, "protected", encoded, decoded)
	compareJOSEField(t, "tag", encoded, decoded)

	// TODO: `ipldNodeToGoPrimitive` doesn't currently work for IPLD nodes with (nested) `bytes` type fields
	//compareJOSEField(t, "recipients", encoded, decoded)
	//compareJOSEField(t, "signatures", encoded, decoded)
	//compareJOSEField(t, "unprotected", encoded, decoded)
}

func compareJOSEField(t *rapid.T, key string, encoded datamodel.Node, decoded datamodel.Node) {
	if encodedField, err := lookupIgnoreNoSuchField(key, encoded); err != nil {
		t.Errorf("error fetching encoded field: %v", err)
	} else if decodedField, err := lookupIgnoreNoSuchField(key, decoded); err != nil {
		t.Errorf("error fetching decoded field: %v", err)
	} else if (encodedField == nil) != (decodedField == nil) {
		t.Errorf("fields must both be present or both be absent:\nencoded{%t}, decoded{%t}", encodedField == nil, decodedField == nil)
	} else if encodedField != nil {
		if (encodedField.Kind() == decodedField.Kind()) ||
			(encodedField.Kind() == datamodel.Kind_Bytes && decodedField.Kind() == datamodel.Kind_String) ||
			(encodedField.Kind() == datamodel.Kind_String && decodedField.Kind() == datamodel.Kind_Bytes) {
			compareNodes(t, encodedField.Kind(), encodedField, decodedField)
		} else {
			t.Errorf("fields must be of the same or compatible kind:\nencoded{%s}\ndecoded{%s}", encodedField.Kind(), decodedField.Kind())
		}
	}
}

func compareNodes(t *rapid.T, kind datamodel.Kind, f1 datamodel.Node, f2 datamodel.Node) {
	var goF1, goF2 interface{}
	switch kind {
	case datamodel.Kind_List:
		goF1 = []interface{}{}
		goF2 = []interface{}{}
	case datamodel.Kind_Map:
	case datamodel.Kind_Link:
		goF1 = map[string]interface{}{}
		goF2 = map[string]interface{}{}
	default:
		compareJOSEBytes(t, f1, f2)
		return
	}
	if err := ipldNodeToGoPrimitive(f1, &goF1); err != nil {
		t.Errorf("error converting field: %v/%v", f1, err)
	} else if err := ipldNodeToGoPrimitive(f2, &goF2); err != nil {
		t.Errorf("error converting field: %v/%v", f2, err)
	} else if !reflect.DeepEqual(goF1, goF2) {
		t.Errorf("fields do not match:\n%s\n%s", goF1, goF2)
	}
}

func compareJOSEBytes(t *rapid.T, f1 datamodel.Node, f2 datamodel.Node) {
	if f1String, err := stringOrBytesAsString(f1); err != nil {
		t.Errorf("error fetching field: %v", err)
	} else if f2String, err := stringOrBytesAsString(f2); err != nil {
		t.Errorf("error fetching field: %v", err)
	} else if f1String != f2String {
		t.Errorf("fields do not match:\n%s\n%s", f1String, f2String)
	}
}

func stringOrBytesAsString(n datamodel.Node) (string, error) {
	if s, err := n.AsString(); err != nil {
		if e, wrongKind := err.(datamodel.ErrWrongKind); wrongKind && (e.ActualKind == datamodel.Kind_Bytes) {
			if b, err := n.AsBytes(); err != nil {
				return "", err
			} else {
				return string(b), nil
			}
		}
		return "", err
	} else {
		return s, nil
	}
}

func TestRoundTripValidJWS(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		validJws := validJWSGen().Draw(t, "valid JWS").(datamodel.Node)
		if roundTrippedJws, err := roundTripJose(validJws); err != nil {
			t.Errorf("failed roundtrip: %v", err)
		} else {
			compareJOSE(t, validJws, roundTrippedJws)
		}
	})
}

// Check that if we encode and decode an arbitrary JOSE object then the
// output is equal to the input (up to ipld normalization)
func TestRoundTripArbitraryJOSE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jose := arbitraryJoseGen().Draw(t, "An arbitrary JOSE object").(datamodel.Node)
		if roundTrippedJose, err := roundTripJose(jose); err != nil {
			t.Errorf("failed roundtrip: %v", err)
		} else {
			compareJOSE(t, jose, roundTrippedJose)
		}
	})
}

// Decoding should always return either a JWS or a JWE if the input is valid
func TestAlwaysDeserializesToEitherJWSOrJWE(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jose := arbitraryJoseGen().Draw(t, "An arbitrary JOSE object").(datamodel.Node)
		if roundTrippedJose, err := roundTripJose(jose); err != nil {
			t.Errorf("failed roundtrip: %v", err)
		} else if jwe, err := isJWE(roundTrippedJose); err != nil {
			t.Errorf("JWE check failed: %v", err)
		} else if !jwe {
			if jws, err := isJWS(roundTrippedJose); err != nil {
				t.Errorf("JWS check failed: %v", err)
			} else {
				require.True(t, jws)
			}
		}
	})
}

// A JWS without a signature is not valid
func TestMissingPayloadErrorParsingJWS(t *testing.T) {
	jsonStr := "{\"signatures\": []}"
	jws, err := parseJOSE([]byte(jsonStr))
	require.NotNil(t, err)
	require.Nil(t, jws)
}

// A JWE without ciphertext is not valid
func TestMissingCiphertextErrorParsingJWE(t *testing.T) {
	jsonStr := "{\"header\": {}}"
	jwe, err := parseJOSE([]byte(jsonStr))
	require.NotNil(t, err)
	require.Nil(t, jwe)
}

// Trying to serialize a JWS with more than one signature to a flattened serialization should throw an error
func TestFlattenedJWSErrorIfSignatureAndSignaturesDefined(t *testing.T) {
	payload := encodeBase64Url(createCid([]byte("payload")).Bytes())
	jsonStr := "{\"signature\": \"sig\", \"signatures\": [], \"payload\": \"" + payload + "\"}"
	jws, err := parseJOSE([]byte(jsonStr))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid JWS serialization")
	require.Nil(t, jws)
}

// If the incoming IPLD data contains a payload which is not a valid CID we
// should raise an error
func TestLoadingJWSWithNonCIDPayloadReturnsError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		payload := nonNilSliceOfBytes().Filter(func(payloadBytes []byte) bool {
			_, _, err := cid.CidFromBytes(payloadBytes)
			return err != nil
		}).Draw(t, "A slice of bytes which is not a valid CID")
		node := fluent.MustBuildMap(
			basicnode.Prototype.Map,
			2,
			func(ma fluent.MapAssembler) {
				ma.AssembleEntry("payload").AssignString(string(payload))
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
		if _, err := storeJOSE(
			ipld.LinkContext{},
			node,
			ls,
		); err != nil {
			require.NotNil(t, err)
			require.Contains(t, err.Error(), "payload is not a valid CID")
		}
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
		jwe, err := parseJOSE(scenario)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "invalid JWE serialization")
		require.Nil(t, jwe)
	}
}
