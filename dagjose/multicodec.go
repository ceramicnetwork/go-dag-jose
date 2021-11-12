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

// Decode deserializes data from the given io.Reader and feeds it into the given datamodel.NodeAssembler. Decode fits
// the codec.Decoder function interface.
func Decode(na datamodel.NodeAssembler, r io.Reader) error {
	// If the passed `NodeAssembler` is not of type `_JOSE__ReprBuilder`, create and use a `_JOSE__ReprBuilder`.
	joseBuilder, alreadyJose := na.(*_JOSE__ReprBuilder)
	if !alreadyJose {
		joseBuilder = Type.JOSE__Repr.NewBuilder().(*_JOSE__ReprBuilder)
	}
	// CBOR is a superset of DAG-JOSE and can be used to decode valid DAG-JOSE objects:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html
	err := cbor.Decode(joseBuilder, r)
	if err != nil {
		return err
	}
	// If `payload` is present but `link` is not, add `link` with the corresponding encoded CID.
	payloadNode := &joseBuilder.w.payload
	linkNode := &joseBuilder.w.link
	if payloadNode.Exists() && !linkNode.Exists() {
		link, err := Type.Base64Url.Link(&payloadNode.v)
		if err != nil {
			return err
		}
		linkNode.m = schema.Maybe_Value
		linkNode.v = _Link{link}
	}
	// The "representation" node gives an accurate view of fields that are actually present
	joseNode := joseBuilder.Build().(schema.TypedNode).Representation()
	if err != nil {
		return err
	}
	// If the passed `NodeAssembler` is not of type `_JOSE__ReprBuilder`, copy the constructed `_JOSE__Repr` into the
	// caller's `NodeAssembler`.
	if !alreadyJose {
		return datamodel.Copy(joseNode, na)
	}
	return nil
}

// Encode walks the given datamodel.Node and serializes it to the given io.Writer. Encode fits the codec.Encoder
// function interface.
func Encode(n datamodel.Node, w io.Writer) error {
	// If the passed `Node` is already of type `_JOSE__Repr`, skip conversion to the latter type.
	_, alreadyJose := n.(*_JOSE__Repr)
	rebuildRequired := false
	if alreadyJose {
		linkNode, err := n.LookupByString("link")
		if err != nil {
			// It's ok for `link` to be absent (even if `payload` was present), but if some other error occurred,
			// return it.
			if _, linkNotFound := err.(datamodel.ErrNotExists); !linkNotFound {
				return err
			}
		} else {
			payloadNode, err := n.LookupByString("payload")
			// If `link` was present then `payload` must be present and the two must match. If any error occurs here
			// (including `payload` being absent) return it.
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
			// The node needs to be rebuilt without `link` before it can be encoded
			rebuildRequired = true
		}
	}
	if !alreadyJose || rebuildRequired {
		joseBuilder := Type.JOSE__Repr.NewBuilder().(*_JOSE__ReprBuilder)
		// Copy the passed `Node` into `_JOSE__ReprBuilder`, which applies all the necessary validations required to
		// construct a `_JOSE__Repr` node.
		if err := datamodel.Copy(n, joseBuilder); err != nil {
			return err
		}
		// Mark `link` as absent because we do not want to encode it
		joseBuilder.w.link.m = schema.Maybe_Absent
		joseBuilder.w.link.v.x = nil
		// The "representation" node gives an accurate view of fields that are actually present
		n = joseBuilder.Build().(schema.TypedNode).Representation()
	}
	// CBOR is a superset of DAG-JOSE and can be used to encode valid DAG-JOSE objects:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html
	return cbor.Encode(n, w)
}
