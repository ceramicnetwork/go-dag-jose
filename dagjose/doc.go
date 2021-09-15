// Package dagjose is an implementation of the IPLD codec.
//
// Data that is encoded using the dag-jose multicodec is guaranteed to be a CBOR
// encoding of the general serialization of either a JWE or a JWS. In order to
// access this information using go-ipld-prime we need two things:
//
// * We need to register an encoder and a decoder which will be used by the cidlink package to decode the raw data into the IPLD data model
//
// * An implementation of ipld.NodeAssembler which knows how to interpret the IPLD data into some concrete go data type which implements ipld.Node
//
// The first of these points is handled by importing this package. There is a side
// effecting operation in the module initialization which registers the encoder
// and decoder with go-ipld-prime.
//
// The latter point is provided by the dagjose.DagJOSE data type. This type
// represents the union of the dagjose.DagJWS and dagjose.DagJWE types.
// Typically, you will use dagjose.LoadJOSE(..) to load a dagjose.DagJOSE
// object, then you will use DagJOSE.AsJWS and DagJOSE.AsJWE to determine
// whether you have a JWS or JWE object respectively.
//
// This package does not provide any direct access to the fields of the JWS and
// JWE objects, instead each kind of JOSE object has a GeneralJSONSerialization
// method which can be used to obtain the general json serialization to be
// passed to third party libraries.
package dagjose
