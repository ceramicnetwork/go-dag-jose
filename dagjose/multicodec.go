package dagjose

import (
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	multicodec "github.com/ipld/go-ipld-prime/multicodec"
)

func init() {
	multicodec.RegisterDecoder(0x85, dagcbor.Decode)
	multicodec.RegisterEncoder(0x85, dagcbor.Encode)
}
