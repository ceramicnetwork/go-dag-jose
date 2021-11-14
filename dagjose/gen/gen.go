package main

import (
	"fmt"
	"os"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
	gengo "github.com/ipld/go-ipld-prime/schema/gen/go"
)

func main() {
	ts := schema.TypeSystem{}
	ts.Init()

	// -- Common types -->

	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnBytes("Bytes"))
	ts.Accumulate(schema.SpawnInt("Int"))
	ts.Accumulate(schema.SpawnFloat("Float"))
	ts.Accumulate(schema.SpawnLink("Link"))
	ts.Accumulate(schema.SpawnMap("Map", "String", "Any", false))
	ts.Accumulate(schema.SpawnList("List", "Any", false))

	// The `Any` union represents a wildcard nested type that can contain any type of scalar or recursive information
	// including itself (as map values or list elements).
	ts.Accumulate(schema.SpawnUnion("Any",
		[]schema.TypeName{
			"String",
			"Bytes",
			"Int",
			"Float",
			"Map",
			"List",
		},
		schema.SpawnUnionRepresentationKinded(map[datamodel.Kind]schema.TypeName{
			datamodel.Kind_String: "String",
			datamodel.Kind_Bytes:  "Bytes",
			datamodel.Kind_Int:    "Int",
			datamodel.Kind_Float:  "Float",
			datamodel.Kind_Map:    "Map",
			datamodel.Kind_List:   "List",
		}),
	))

	// -- Decode types -->

	// While `Base64Url` is a `String` type and generated through the schema, it has some (surgical) modifications that
	// allow it to be treated as a base64url-encoded string "lens" looking at raw, un-encoded bytes being decoded.
	ts.Accumulate(schema.SpawnString("Base64Url"))

	// JWS
	ts.Accumulate(schema.SpawnStruct("DecodedSignature", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("protected", "Base64Url", true, false),
		schema.SpawnStructField("signature", "Base64Url", false, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("DecodedSignatures", "DecodedSignature", false))

	// JWE
	ts.Accumulate(schema.SpawnStruct("DecodedRecipient", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("encrypted_key", "Base64Url", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("DecodedRecipients", "DecodedRecipient", false))

	// JOSE
	ts.Accumulate(schema.SpawnStruct("DecodedJOSE", []schema.StructField{
		schema.SpawnStructField("aad", "Base64Url", true, false),
		schema.SpawnStructField("ciphertext", "Base64Url", true, false),
		schema.SpawnStructField("iv", "Base64Url", true, false),
		// `link` is not encoded as part of DAG-JOSE because it is not included in the DAG-JOSE spec but is included
		// here in the schema because it is required when decoding/encoding from/to other encodings (e.g. DAG-JSON).
		// If `payload` is present during decode, `link` is added with contents matching `payload`. If `link` is present
		// during encode, it is validated against `payload` and then ignored.
		schema.SpawnStructField("link", "Link", true, false),
		schema.SpawnStructField("payload", "Base64Url", true, false),
		schema.SpawnStructField("protected", "Base64Url", true, false),
		schema.SpawnStructField("recipients", "DecodedRecipients", true, false),
		schema.SpawnStructField("signatures", "DecodedSignatures", true, false),
		schema.SpawnStructField("tag", "Base64Url", true, false),
		schema.SpawnStructField("unprotected", "Any", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	// -- Encode types -->

	// While `Raw` is a `Bytes` type and generated through the schema, it has some (surgical) modifications that allow
	// it to be treated as a raw, un-encoded bytes "lens" looking at base64url-encoded strings being encoded.
	ts.Accumulate(schema.SpawnBytes("Raw"))

	// JWS
	ts.Accumulate(schema.SpawnStruct("EncodedSignature", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("protected", "Raw", true, false),
		schema.SpawnStructField("signature", "Raw", false, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("EncodedSignatures", "EncodedSignature", false))

	// JWE
	ts.Accumulate(schema.SpawnStruct("EncodedRecipient", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("encrypted_key", "Raw", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("EncodedRecipients", "EncodedRecipient", false))

	// JOSE
	ts.Accumulate(schema.SpawnStruct("EncodedJOSE", []schema.StructField{
		schema.SpawnStructField("aad", "Raw", true, false),
		schema.SpawnStructField("ciphertext", "Raw", true, false),
		schema.SpawnStructField("iv", "Raw", true, false),
		// `link` is not encoded as part of DAG-JOSE because it is not included in the DAG-JOSE spec but is included
		// here in the schema because it is required when decoding/encoding from/to other encodings (e.g. DAG-JSON).
		// If `payload` is present during decode, `link` is added with contents matching `payload`. If `link` is present
		// during encode, it is validated against `payload` and then ignored.
		schema.SpawnStructField("link", "Link", true, false),
		schema.SpawnStructField("payload", "Raw", true, false),
		schema.SpawnStructField("protected", "Raw", true, false),
		schema.SpawnStructField("recipients", "EncodedRecipients", true, false),
		schema.SpawnStructField("signatures", "EncodedSignatures", true, false),
		schema.SpawnStructField("tag", "Raw", true, false),
		schema.SpawnStructField("unprotected", "Any", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	if errs := ts.ValidateGraph(); errs != nil {
		for _, err := range errs {
			fmt.Printf("- %s\n", err)
		}
		panic("invalid schema")
	}

	gengo.Generate(os.Args[1], "dagjose", ts, &gengo.AdjunctCfg{
		// This is important for the `Any` union to work correctly
		CfgUnionMemlayout: map[schema.TypeName]string{"Any": "interface"},
	})
}
