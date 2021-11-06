package main

import (
	"fmt"
	"os"

	"github.com/ipld/go-ipld-prime/schema"
	gengo "github.com/ipld/go-ipld-prime/schema/gen/go"
)

func main() {
	ts := schema.TypeSystem{}
	ts.Init()

	// Common
	ts.Accumulate(schema.SpawnString("String"))
	ts.Accumulate(schema.SpawnMap("Map", "String", "String", false))

	// JWS
	ts.Accumulate(schema.SpawnStruct("Signature", []schema.StructField{
		schema.SpawnStructField("header", "Map", true, false),
		schema.SpawnStructField("protected", "String", true, false),
		schema.SpawnStructField("signature", "String", false, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	ts.Accumulate(schema.SpawnList("Signatures", "Signature", false))

	// JWE
	ts.Accumulate(schema.SpawnStruct("Recipient", []schema.StructField{
		schema.SpawnStructField("header", "Map", true, false),
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
		schema.SpawnStructField("unprotected", "Map", true, false),
	}, schema.SpawnStructRepresentationMap(nil)))

	if errs := ts.ValidateGraph(); errs != nil {
		for _, err := range errs {
			fmt.Printf("- %s\n", err)
		}
		panic("not happening")
	}

	gengo.Generate(os.Args[1], "dagjose", ts, &gengo.AdjunctCfg{})
}
