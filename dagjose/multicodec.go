package dagjose

//go:generate go run ./gen .
//go:generate go fmt ./

import (
	"errors"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/codec"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multibase"
)

func init() {
	multicodec.RegisterDecoder(0x85, Decode)
	multicodec.RegisterEncoder(0x85, Encode)
}

// Decode deserializes data from the given io.Reader and feeds it into the given datamodel.NodeAssembler. Decode fits
// the codec.Decoder function interface.
func Decode(na datamodel.NodeAssembler, r io.Reader) error {
	// If the passed `NodeAssembler` is not of type `_DecodedJOSE__ReprBuilder`, create and use a
	// `_DecodedJOSE__ReprBuilder`.
	joseBuilder, alreadyJose := na.(*_DecodedJOSE__ReprBuilder)
	if !alreadyJose {
		joseBuilder = Type.DecodedJOSE__Repr.NewBuilder().(*_DecodedJOSE__ReprBuilder)
	}
	// DAG-CBOR is a superset of DAG-JOSE and can be used to decode valid DAG-JOSE objects. Use DAG-CBOR decoding but do
	// not allow IPLD Links. See: https://specs.ipld.io/block-layer/codecs/dag-jose.html
	err := dagcbor.DecodeOptions{
		AllowLinks: false,
	}.Decode(joseBuilder, r)
	if err != nil {
		return err
	}
	// If `payload` is present but `link` is not, add `link` with the corresponding encoded CID.
	payloadNode := &joseBuilder.w.payload
	linkNode := &joseBuilder.w.link
	if payloadNode.Exists() && !linkNode.Exists() {
		if link, err := Type.Base64Url.Link(&payloadNode.v); err != nil {
			return err
		} else {
			linkNode.m = schema.Maybe_Value
			linkNode.v = *link
		}
	}
	// The "representation" node gives an accurate view of fields that are actually present
	joseNode := joseBuilder.Build().(schema.TypedNode).Representation()
	// If the passed `NodeAssembler` is not of type `_DecodedJOSE__ReprBuilder`, copy the constructed
	// `_DecodedJOSE__Repr` into the caller's `NodeAssembler`.
	if !alreadyJose {
		return datamodel.Copy(joseNode, na)
	}
	return nil
}

// Encode walks the given datamodel.Node and serializes it to the given io.Writer. Encode fits the codec.Encoder
// function interface.
func Encode(n datamodel.Node, w io.Writer) error {
	if n, err := sanitizeNodeForEncode(n); err != nil {
		return err
	} else {
		// DAG-CBOR is a superset of DAG-JOSE and can be used to encode valid DAG-JOSE objects. Use DAG-CBOR's Map sorting
		// but do not allow IPLD Links. See: https://specs.ipld.io/block-layer/codecs/dag-jose.html
		return dagcbor.EncodeOptions{
			AllowLinks:  false,
			MapSortMode: codec.MapSortMode_RFC7049,
		}.Encode(n, w)
	}
}

func sanitizeNodeForEncode(n datamodel.Node) (datamodel.Node, error) {
	rebuildRequired := false
	// If `link` and `payload` are present, make sure they match.
	if linkNode, err := n.LookupByString("link"); err != nil {
		// It's ok for `link` to be absent (even if `payload` was present), but if some other error occurred,
		// return it.
		if _, linkNotFound := err.(datamodel.ErrNotExists); !linkNotFound {
			return nil, err
		}
	} else {
		// If `link` was present then `payload` must be present and the two must match. If any error occurs here
		// (including `payload` being absent) return it.
		payloadNode, err := n.LookupByString("payload")
		if err != nil {
			return nil, err
		}
		payloadString, err := payloadNode.AsString()
		if err != nil {
			return nil, err
		}
		cidFromPayload, err := cid.Decode(string(multibase.Base64url) + payloadString)
		if err != nil {
			return nil, err
		}
		linkFromNode, err := linkNode.AsLink()
		if err != nil {
			return nil, err
		}
		if linkFromNode.(cidlink.Link).Cid != cidFromPayload {
			return nil, errors.New("cid mismatch")
		}
		// The node needs to be rebuilt without `link` before it can be encoded
		rebuildRequired = true
	}
	// If the passed `Node` is not of type `_EncodedJOSE__Repr`, convert it to `_EncodedJOSE__Repr`.
	if _, alreadyJose := n.(*_EncodedJOSE__Repr); !alreadyJose || rebuildRequired {
		joseBuilder := Type.EncodedJOSE__Repr.NewBuilder().(*_EncodedJOSE__ReprBuilder)
		// Copy the passed `Node` into `_EncodedJOSE__ReprBuilder`, which applies all the necessary validations required
		// to construct a `_EncodedJOSE__Repr` node.
		if err := datamodel.Copy(n, joseBuilder); err != nil {
			return nil, err
		}
		// Mark `link` as absent because we do not want to encode it
		joseBuilder.w.link.m = schema.Maybe_Absent
		joseBuilder.w.link.v.x = nil
		// The "representation" node gives an accurate view of fields that are actually present
		n = joseBuilder.Build().(schema.TypedNode).Representation()
	}
	return n, nil
}
