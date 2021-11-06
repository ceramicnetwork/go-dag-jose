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

	// Common types
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnBytes("Bytes"))
	ts.Accumulate(schema.SpawnInt("Int"))
	ts.Accumulate(schema.SpawnFloat("Float"))
	ts.Accumulate(schema.SpawnMap("Map", "String", "Any", false))
	ts.Accumulate(schema.SpawnList("List", "Any", false))

	// The `Any` union represents a wildcard nested type that can contain any
	// type of information including itself (as map values or list elements).
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

	// JWS
	ts.Accumulate(schema.SpawnStruct("Signature", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("protected", "String", true, false),
		schema.SpawnStructField("signature", "String", false, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("Signatures", "Signature", false))

	// JWE
	ts.Accumulate(schema.SpawnStruct("Recipient", []schema.StructField{
		schema.SpawnStructField("header", "Any", true, false),
		schema.SpawnStructField("encrypted_key", "String", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("Recipients", "Recipient", false))

	// JOSE
	ts.Accumulate(schema.SpawnStruct("JOSE", []schema.StructField{
		schema.SpawnStructField("aad", "String", true, false),
		schema.SpawnStructField("ciphertext", "String", true, false),
		schema.SpawnStructField("iv", "String", true, false),
		schema.SpawnStructField("payload", "String", true, false),
		schema.SpawnStructField("protected", "String", true, false),
		schema.SpawnStructField("recipients", "Recipients", true, false),
		schema.SpawnStructField("signatures", "Signatures", true, false),
		schema.SpawnStructField("tag", "String", true, false),
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
