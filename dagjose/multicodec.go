package dagjose

//go:generate go run ./gen .
//go:generate go fmt ./

import (
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/codec/cbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/schema"
)

func init() {
	multicodec.RegisterDecoder(0x85, Decode)
	multicodec.RegisterEncoder(0x85, Encode)
}

// Decode deserializes data from the given io.Reader and feeds it into the
// given datamodel.NodeAssembler. Decode fits the codec.Decoder function
// interface.
func Decode(na datamodel.NodeAssembler, r io.Reader) error {
	// If the passed `NodeAssembler` is not of type `_JOSE__ReprBuilder`, create
	// and use a `_JOSE__ReprBuilder` to decode.
	nodeBuilder, alreadyJose := na.(*_JOSE__ReprBuilder)
	if !alreadyJose {
		nodeBuilder = Type.JOSE__Repr.NewBuilder().(*_JOSE__ReprBuilder)
	}
	// CBOR is a superset of (DAG-)JOSE and can be used to decode valid
	// DAG-JOSE objects without decoding the CID (as expected by the
	// DAG-JOSE spec:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
	err := cbor.Decode(nodeBuilder, r)
	if err != nil {
		return err
	}
	// If `payload` is present but `link` is not, add `link` with the
	// corresponding encoded CID.
	payloadNode := &nodeBuilder.w.payload
	linkNode := &nodeBuilder.w.link
	if payloadNode.Exists() && !linkNode.Exists() {
		link, err := Type.Base64Url.Link(&payloadNode.v)
		if err != nil {
			return err
		}
		linkNode.m = schema.Maybe_Value
		linkNode.v = _Link{link}
	}
	// The "representation" node gives an accurate view of fields that are
	// actually present.
	n := nodeBuilder.Build().(schema.TypedNode).Representation()
	// If the passed `NodeAssembler` is not of type `_JOSE__ReprBuilder`, copy
	// the constructed `_JOSE__Repr` into the caller's `NodeAssembler`.
	if !alreadyJose {
		return datamodel.Copy(n, na)
	}
	return nil
}

// Encode walks the given datamodel.Node and serializes it to the given
// io.Writer. Encode fits the codec.Encoder function interface.
func Encode(n datamodel.Node, w io.Writer) error {
	// If the passed `Node` is already of type `_JOSE__Repr`, skip conversion to
	// the latter type.
	if _, ok := n.(*_JOSE__Repr); !ok {
		// Use `datamodel.Copy` to convert the passed `Node` into `_JOSE__Repr`,
		// which applies all the necessary validations required to construct a
		// proper DAG-JOSE Node.
		dagJOSEBuilder := Type.JOSE__Repr.NewBuilder()
		if err := datamodel.Copy(n, dagJOSEBuilder); err != nil {
			return err
		}
		linkNode, err := n.LookupByString("link")
		if linkNode == nil {
			// It's ok for `link` to be absent, but if some other error occurred, return it.
			if _, linkNotFound := err.(datamodel.ErrNotExists); !linkNotFound {
				return err
			}
		} else {
			payloadNode, err := n.LookupByString("payload")
			// If `link` was present then `payload` must be present and the two must match. If any error occurs here
			// (including `payload` being missing) return it.
			if err != nil {
				return err
			}
			payloadString, err := payloadNode.AsString()
			if err != nil {
				return err
			}
			cidFromPayload, err := cid.Decode("u" + payloadString)
			if err != nil {
				return err
			}
			linkFromPayload := cidlink.Link{Cid: cidFromPayload}
			linkFromNode, err := linkNode.AsLink()
			if err != nil {
				return err
			}
			if linkFromPayload != linkFromNode {
				return cid.ErrCidTooShort
			}
			// Mark `link` as absent because we do not want to encode it
			dagJOSEBuilder.(*_JOSE__ReprBuilder).w.link.m = schema.Maybe_Absent
		}
		n = dagJOSEBuilder.Build().(schema.TypedNode).Representation()
	}
	// CBOR is a superset of (DAG-)JOSE and can be used to encode valid DAG-JOSE
	// objects without encoding the CID (as expected by the DAG-JOSE spec:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
	return cbor.Encode(n, w)
}
