package dagjose

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	cbor "github.com/fxamacker/cbor/v2"
	ipld "github.com/ipld/go-ipld-prime"
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func init() {
	cidlink.RegisterMulticodecDecoder(0x85, Decoder)
	cidlink.RegisterMulticodecEncoder(0x85, dagcbor.Encoder)
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
			Payload    []byte `cbor:"payload"`
			Signatures []struct {
				Signature []byte `cbor:"signature"`
				Protected []byte `cbor:"protected"`
			} `cbor:"signatures"`
		}
		decoder := cbor.NewDecoder(bytes.NewReader(rawData))
		err = decoder.Decode(&rawDecoded)
		if err != nil {
			return fmt.Errorf("error decoding CBOR for dag-jose: %v", err)
		}
		joseAssembler.dagJose.payload = rawDecoded.Payload
		for _, sig := range rawDecoded.Signatures {
			joseAssembler.dagJose.signatures = append(
				joseAssembler.dagJose.signatures,
				JOSESignature{
					protected: sig.Protected,
					signature: sig.Signature,
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
