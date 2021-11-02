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
	ts.Accumulate(schema.SpawnBytes("Bytes"))
	//ts.Accumulate(schema.SpawnMap("Map", "String", "String", false))

	// JWS
	ts.Accumulate(schema.SpawnStruct("Signature", []schema.StructField{
		//schema.SpawnStructField("header", "Map", true, false),
		schema.SpawnStructField("protected", "Bytes", true, false),
		schema.SpawnStructField("signature", "Bytes", false, false),
	}, schema.SpawnStructRepresentationMap(map[string]string{})))

	ts.Accumulate(schema.SpawnList("Signatures", "Signature", false))

	// JWE
	ts.Accumulate(schema.SpawnStruct("Recipient", []schema.StructField{
		//schema.SpawnStructField("header", "Map", true, false),
		schema.SpawnStructField("encrypted_key", "Bytes", true, false),
	}, schema.SpawnStructRepresentationMap(map[string]string{})))

	ts.Accumulate(schema.SpawnList("Recipients", "Recipient", false))

	// JOSE
	ts.Accumulate(schema.SpawnStruct("JOSE", []schema.StructField{
		schema.SpawnStructField("aad", "Bytes", true, false),
		schema.SpawnStructField("ciphertext", "Bytes", true, false),
		schema.SpawnStructField("iv", "Bytes", true, false),
		schema.SpawnStructField("payload", "Bytes", true, false),
		schema.SpawnStructField("protected", "Bytes", true, false),
		schema.SpawnStructField("recipients", "Recipients", true, false),
		schema.SpawnStructField("signatures", "Signatures", true, false),
		schema.SpawnStructField("tag", "Bytes", true, false),
		//schema.SpawnStructField("unprotected", "Map", true, false),
	}, schema.SpawnStructRepresentationMap(map[string]string{})))

	if errs := ts.ValidateGraph(); errs != nil {
		for _, err := range errs {
			fmt.Printf("- %s\n", err)
		}
		panic("not happening")
	}

	gengo.Generate(os.Args[1], "dagjose", ts, &gengo.AdjunctCfg{})
}
