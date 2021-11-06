package dagjose

//go:generate go run ./gen .
//go:generate go fmt ./

import (
	"io"

	"github.com/ipld/go-ipld-prime/codec/cbor"
	"github.com/ipld/go-ipld-prime/datamodel"
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
	// one of the latter type, use it, then copy the constructed `_JOSE__Repr`
	// into the caller's `NodeAssembler`.
	if _, ok := na.(*_JOSE__ReprBuilder); ok {
		return cbor.Decode(na, r)
	} else {
		dagJOSEBuilder := Type.JOSE__Repr.NewBuilder()
		// CBOR is a superset of (DAG-)JOSE and can be used to decode valid
		// DAG-JOSE objects without decoding the CID (as expected by the
		// DAG-JOSE spec:
		// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
		err := cbor.Decode(dagJOSEBuilder, r)
		if err != nil {
			return err
		}
		n := dagJOSEBuilder.Build().(schema.TypedNode).Representation()
		return datamodel.Copy(n, na)
	}
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
		err := datamodel.Copy(n, dagJOSEBuilder)
		if err != nil {
			return err
		}
		n = dagJOSEBuilder.Build().(schema.TypedNode).Representation()
	}
	// CBOR is a superset of (DAG-)JOSE and can be used to encode valid DAG-JOSE
	// objects without encoding the CID (as expected by the DAG-JOSE spec:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
	return cbor.Encode(n, w)
}
