package go_dag_jose

import (
	"fmt"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/alexjg/go-dag-jose/dagjose"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func init() {
	multicodec.RegisterDecoder(0x85, dagcbor.Decode)
	multicodec.RegisterEncoder(0x85, dagcbor.Encode)
}

func TestRead(t *testing.T) {
	// Here we're creating a `CID` which points to a JWS
	jwsCID, err := cid.Decode("bafkrkmgrx7mmc6o4mtx27ajzc4pp556jm4fapndguch3r2457gdhgdmjr3kd3n2uxoaqwsdjvusniyb7saza")
	assert.NoError(t, err)

	// cidlink.Link is an implementation of `ipld.Link` backed by a CID
	jwsLnk := cidlink.Link{Cid: jwsCID}

	ls := cidlink.DefaultLinkSystem()

	jose, err := dagjose.LoadJOSE(
		jwsLnk,
		ipld.LinkContext{},
		ls, //<an implementation of ipld.Loader, which knows how to get the block data from IPFS>,
	)
	assert.NoError(t, err)

	asJWS := jose.AsJWS()
	assert.NotEmpty(t, asJWS)

	// We have a JWS object, print the general serialization of it
	fmt.Print(jose.AsJWS().GeneralJSONSerialization())
}

func TestWrite(t *testing.T) {
	jws := "{\"payload\":\"AVUVMERzfWva7RtEDmEapvAFLHhJsyKwB2A93ecOaNsx9mGVr0CpPoGbZY99Ko4dc8_QYA\",\"signatures\":[{\"protected\":\"AQW1Dw\",\"signature\":\"_w\"},{\"header\":{\"\":[false,\"Ⅴ𐅕~!⅋:\",[null,[0],{\"\":0,\"$ ‮\\u001bᾉ~�ǂ\\t~\\t~¢?̳\":\"\",\"$ऻ꙱Ɖ\u007F![  \\u003c!ᰩ\":\"~#\\u0000-\",\"*\\u003c:𝒌Œ\\u0000﮶`+?ો\":false,\"=۳-\\u0026;ð\":9846576,\"°\":null,\"ǅ/ࢥ?$ˎʱ$?!ɔ ~\":-2844,\"Ⱥ𐧚\\u003c\":null,\"\uE23Fڈ͎Ⅸ@൙~\":false,\"\uECAD\":\"~-ᵊ\",\"\U0010E48E\\u0019𖽙\":-1.0452552527236051e-299},{\"'\":\"ৗ\",\"/\\u000b\":null,\"@\\u0026_\":-45005838,\"\\\\\uE007#🄌#!a\\u003cᶟǋ\u007F́;a\":-2.477489147167944e-8,\"�ꙶ\u007F\\\"\":\"?©₷ઉ𝞶Ⱥ৺؀\"},false,-0.14566001789073013,7567,-356.56302885114513],\"𝘣~ǅ!\\u001b̎ऻ?\"],\"![!\U0001ECB0:Ⱥ v~=Y~7.\uE017𞥃$ᾨ꙲\":-34913,\"q\":null,\"ǂ⃞� ~~ ~\":true,\"\uE000\uE001\\\"𑢢Þ̽$˔\":\"歚\"},\"protected\":\"Pw\",\"signature\":\"BcwqFXUP4FA\"},{\"header\":{\"\":[-1168],\"%~ᾎ[@୳Ⱥ\":null,\".\":-19284},\"signature\":\"4gMAMgjc\"},{\"signature\":\"A7YAnwACHg\"},{\"protected\":\"AXQBWAIAfQEGAlI\",\"signature\":\"AQ\"},{\"signature\":\"ZQE\"},{\"protected\":\"Mwo\",\"signature\":\"cgArAQ\"}]}"
	dagJWS, err := dagjose.ParseJWS([]byte(jws))
	assert.NoError(t, err)

	ls := cidlink.DefaultLinkSystem()
	link, err := dagjose.StoreJOSE(
		ipld.LinkContext{},
		dagJWS.AsJOSE(),
		ls,
	)
	assert.NoError(t, err)

	fmt.Printf("Link is: %v", link)
}
