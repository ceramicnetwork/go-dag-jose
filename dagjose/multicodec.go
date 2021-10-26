package dagjose

import (
	"io"
	"reflect"

	"github.com/ipld/go-ipld-prime/codec/cbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/multicodec"
)

func init() {
	multicodec.RegisterDecoder(0x85, Decode)
	multicodec.RegisterEncoder(0x85, Encode)
}

// Decode deserializes data from the given io.Reader and feeds it into the
// given datamodel.NodeAssembler. Decode fits the codec.Decoder function
// interface.
func Decode(na datamodel.NodeAssembler, r io.Reader) error {
	// CBOR is a superset of DAG-JOSE and can be used to decode valid DAG-JOSE
	// objects without decoding the CID (as expected by the DAG-JOSE spec:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
	if reflect.TypeOf(na) == reflect.TypeOf(new(dagJOSENodeBuilder)) {
		return cbor.Decode(na, r)
	} else {
		// If the passed `datamodel.NodeAssembler` is not of type
		// `dagjose.dagJOSENodeBuilder`, create one of the latter type, use it,
		// then copy the constructed `dagjose.dagJOSENode` into the caller's
		// `datamodel.NodeAssembler`.
		dagJoseBuilder := NewBuilder()
		err := cbor.Decode(dagJoseBuilder, r)
		if err != nil {
			return err
		}
		return datamodel.Copy(dagJoseBuilder.Build(), na)
	}
}

// Encode walks the given datamodel.Node and serializes it to the given
// io.Writer. Encode fits the codec.Encoder function interface.
func Encode(n datamodel.Node, w io.Writer) error {
	// If the passed `datamodel.Node` is already of type `dagjose.dagJOSENode`,
	// skip conversion to the latter type.
	if reflect.TypeOf(n) != reflect.TypeOf(new(dagJOSENode)) {
		// Use `datamodel.Copy` to convert the passed `datamodel.Node` into a
		// `dagjose.dagJOSENode` via `dagjose.dagJOSENodeBuilder`, which applies
		// all the necessary validations required to construct a proper DAG-JOSE
		// IPLD Node.
		dagJoseBuilder := NewBuilder()
		err := datamodel.Copy(n, dagJoseBuilder)
		if err != nil {
			return err
		}
		n = dagJoseBuilder.Build()
	}
	// CBOR is a superset of DAG-JOSE and can be used to encode valid DAG-JOSE
	// objects without encoding the CID (as expected by the DAG-JOSE spec:
	// https://specs.ipld.io/block-layer/codecs/dag-jose.html).
	return cbor.Encode(n, w)
}
