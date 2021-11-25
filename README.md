# go-dag-jose

This is an implementation of the IPLD [dag-jose codec](https://ipld.io/specs/codecs/dag-jose/spec/).

Data that is encoded using the `dag-jose` codec is guaranteed to be a CBOR encoding of the general serialization of
either a JWE or a JWS.

Module initialization registers the `dagjose.Encode` and `dagjose.Decode` with `go-ipld-prime`.

## TODOs

- [ ] Add support for "compact" JWE/JWS serialization
- [ ] Add CI pipeline
- [ ] Add support for comparing recursive types in unit tests