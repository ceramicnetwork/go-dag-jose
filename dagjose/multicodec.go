package dagjose

import (
	"io"

	ipld "github.com/ipld/go-ipld-prime"
	dagcbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func init() {
	cidlink.RegisterMulticodecDecoder(0x85, Decoder)
	cidlink.RegisterMulticodecEncoder(0x85, dagcbor.Encoder)
}

func Decoder(na ipld.NodeAssembler, r io.Reader) error {
	err := dagcbor.Decoder(na, r)
	if err != nil {
		return err
	}
	return nil
}
