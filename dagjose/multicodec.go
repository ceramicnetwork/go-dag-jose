package dagjose

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	cbor "github.com/fxamacker/cbor/v2"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func init() {
	cidlink.RegisterMulticodecDecoder(0x85, Decoder)
	cidlink.RegisterMulticodecEncoder(0x85, dagcbor.Encoder)
}

type JOSESignature struct {
	protected *string
	header    *string
	signature []byte
}

type JWERecipient struct {
	header        map[string]string
	encrypted_key *string
}

type DagJOSE struct {
	// JWS top level keys
	payload    *cid.Cid
	signatures []JOSESignature
	// JWE top level keys
	protected   *string
	unprotected *string
	iv          *string
	aad         *string
	ciphertext  *string
	tag         *string
	recipients  []JWERecipient
}

func NewDagJWS(jsonSerialization string) (*DagJOSE, error) {
	var rawJws struct {
		Payload   string `json:"payload"`
		Protected string `json:"protected"`
		Signature string `json:"signature"`
	}
	err := json.Unmarshal([]byte(jsonSerialization), &rawJws)
	if err != nil {
		return nil, fmt.Errorf("error deserializing json: %v", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(rawJws.Payload)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload bytes: %v", err)
	}
	_, payloadCid, err := cid.CidFromBytes(payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("error deserializing payload as a Cid: %v", err)
	}
	signatureBytes, err := base64.RawURLEncoding.DecodeString(rawJws.Signature)
	if err != nil {
		return nil, fmt.Errorf("error decoding signature bytes: %v", err)
	}
	return &DagJOSE{
		payload: &payloadCid,
		signatures: []JOSESignature{
			{
				protected: &rawJws.Protected,
				signature: signatureBytes,
				header:    nil,
			},
		},
	}, nil
}

func (d *DagJOSE) GeneralJSONSerialization() string {
	jsonJose := make(map[string]interface{})
	jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.payload.Bytes())
	sigs := make([]map[string]string, 0)
	for _, sig := range d.signatures {
		jsonSig := make(map[string]string)
		if sig.protected != nil {
			jsonSig["protected"] = *sig.protected
		}
		jsonSig["signature"] = base64.RawURLEncoding.EncodeToString(sig.signature)
		sigs = append(sigs, jsonSig)
	}
	jsonJose["signatures"] = sigs
	encoded, err := json.Marshal(jsonJose)
	if err != nil {
		panic("impossible")
	}
	return string(encoded)
}

func Decoder(na ipld.NodeAssembler, r io.Reader) error {
	joseAssembler, isJoseAssembler := na.(*DagJOSENodeBuilder)
	// THIS IS A HACK
	// Rather than implementing the `NodeAssembler` interface, we are just
	// checking if the user has indicated that they want to construct a
	// DagJOSE, which they do by passing a DagJOSENodeBuilder to ipld.Link.Load.
	// We then proceed to decode the data to CBOR, and then construct a DagJOSE
	// object from the deserialized CBOR. Allocating an intermediary object
	// is explicitly what the whole `NodeAssembler` machinery is designed to avoid
	// so we absolutely should not do this.
	//
	// The next step here is to implement `NodeAssembler` (in `assembler.go`)
	// in such a way that it throws errors if the incoming data does not match
	// the expected layout of a dag-jose object. The only reason I have not
	// done this yet is that it requires a lot of code to implement NodeAssembler
	// and I wanted to check that the user facing API made sense first.
	//
	// This is also why this code contains very little error checking, we'll be
	// doing that more thoroughly in the NodeAssembler implementation
	if isJoseAssembler {
        rawData, err := ioutil.ReadAll(r)
        if err != nil {
            return fmt.Errorf("error reading: %v", err)
        }
        var rawDecoded struct {
            Payload []byte `cbor:"payload"`
            Signatures []struct{
                Signature []byte `cbor:"signature"`
                Protected []byte `cbor:"protected"`
            } `cbor:"signatures"`
        }
		decoder := cbor.NewDecoder(bytes.NewReader(rawData))
		err = decoder.Decode(&rawDecoded)
		if err != nil {
			return fmt.Errorf("error decoding CBOR for dag-jose: %v", err)
		}
        fmt.Printf("Raw decoded: %v\n", rawDecoded)
		cidPayload, err := cid.Cast(rawDecoded.Payload)
		if err != nil {
			return fmt.Errorf("Error casting payload to cid: %v", err)
		}
		joseAssembler.dagJose.payload = &cidPayload
		for _, sig := range rawDecoded.Signatures {
			protected := base64.RawURLEncoding.EncodeToString(sig.Protected)
			signature := sig.Signature
			joseAssembler.dagJose.signatures = append(
				joseAssembler.dagJose.signatures,
				JOSESignature{
					protected: &protected,
					signature: signature,
					header:    nil,
				},
			)
		}
		return nil
	}
	err := dagcbor.Decoder(na, r)
	if err != nil {
		return err
	}
	return nil
}
