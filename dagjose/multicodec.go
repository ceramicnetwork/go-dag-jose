package dagjose

import (
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func init() {
	cidlink.RegisterMulticodecDecoder(0x85, dagcbor.Decoder)
	cidlink.RegisterMulticodecEncoder(0x85, dagcbor.Encoder)
}
