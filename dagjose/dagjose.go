package dagjose

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type JWSSignature struct {
	protected []byte
	header    map[string]ipld.Node
	signature []byte
}

type JWERecipient struct {
	header        map[string]ipld.Node
	encrypted_key []byte
}

// This is a union of the DagJWE and DagJWS types. Typically you will want to
// as AsJWE and AsJWS to get a concrete JOSE object.
type DagJOSE struct {
	// JWS top level keys
	payload    *cid.Cid
	signatures []JWSSignature
	// JWE top level keys
	protected   []byte
	unprotected []byte
	iv          []byte
	aad         []byte
	ciphertext  []byte
	tag         []byte
	recipients  []JWERecipient
}

// If this jose object is a JWS then this will return a DagJWS, if it is a
// JWE then AsJWS will return nil
func (d *DagJOSE) AsJWS() *DagJWS {
	if d.payload != nil {
		return &DagJWS{dagjose: d}
	}
	return nil
}

// If this jose object is a JWE then this will return a DagJWE, if it is a
// JWS then AsJWE will return nil
func (d *DagJOSE) AsJWE() *DagJWE {
	if d.ciphertext != nil {
		return &DagJWE{dagjose: d}
	}
	return nil
}

type DagJWS struct{ dagjose *DagJOSE }

// Returns a DagJOSE object that implements ipld.Node and can be passed to
// ipld related infrastructure
func (d *DagJWS) AsJOSE() *DagJOSE {
	return d.dagjose
}

type DagJWE struct{ dagjose *DagJOSE }

// Returns a DagJOSE object that implements ipld.Node and can be passed to
// ipld related infrastructure
func (d *DagJWE) AsJOSE() *DagJOSE {
	return d.dagjose
}

// Given a JSON string reresenting a JWS in either general or compact serialization this
// will return a DagJWS
func ParseJWS(jsonStr []byte) (*DagJWS, error) {
	var rawJws struct {
		Payload    *string `json:"payload"`
		Signatures []struct {
			Protected *string                `json:"protected"`
			Signature string                 `json:"signature"`
			Header    map[string]interface{} `json:"header"`
		} `json:"signatures"`
		Protected *string                `json:"protected"`
		Signature *string                `json:"signature"`
		Header    map[string]interface{} `json:"header"`
	}
	if err := json.Unmarshal(jsonStr, &rawJws); err != nil {
		return nil, fmt.Errorf("error parsing jws json: %v", err)
	}
	result := DagJOSE{}

	if rawJws.Payload == nil {
		return nil, fmt.Errorf("JWS has no payload property")
	}

	if rawJws.Signature != nil && rawJws.Signatures != nil {
		return nil, fmt.Errorf("JWS JSON cannot contain both a 'signature' and a 'signatures' key")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(*rawJws.Payload)
	if err != nil {
		return nil, fmt.Errorf("error parsing payload: %v", err)
	}
	_, cid, err := cid.CidFromBytes(payloadBytes)
	if err != nil {
		panic(fmt.Errorf("payload is not a CID"))
	}
	result.payload = &cid

	var sigs []JWSSignature
	if rawJws.Signature != nil {
		sig := JWSSignature{}

		sigBytes, err := base64.RawURLEncoding.DecodeString(*rawJws.Signature)
		if err != nil {
			return nil, fmt.Errorf("error decoding signature: %v", err)
		}
		sig.signature = sigBytes

		if rawJws.Protected != nil {
			protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJws.Protected)
			if err != nil {
				return nil, fmt.Errorf("error parsing signature: %v", err)
			}
			sig.protected = protectedBytes
		}

		if rawJws.Header != nil {
			header := make(map[string]ipld.Node)
			for key, v := range rawJws.Header {
				node, err := goPrimitiveToIpldBasicNode(v)
				if err != nil {
					return nil, fmt.Errorf("error converting header value for key '%s'  of to ipld: %v", key, err)
				}
				header[key] = node
			}
			sig.header = header
		}
		sigs = append(sigs, sig)
	} else if rawJws.Signatures != nil {
		sigs = make([]JWSSignature, 0, len(rawJws.Signatures))
		for idx, rawSig := range rawJws.Signatures {
			sig := JWSSignature{}
			if rawSig.Protected != nil {
				protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawSig.Protected)
				if err != nil {
					return nil, fmt.Errorf("error parsing signatures[%d]['protected']: %v", idx, err)
				}
				sig.protected = protectedBytes
			}

			if rawSig.Header != nil {
				header := make(map[string]ipld.Node)
				for key, v := range rawSig.Header {
					node, err := goPrimitiveToIpldBasicNode(v)
					if err != nil {
						return nil, fmt.Errorf("error converting header value for key '%s'  of sign %d to ipld: %v", key, idx, err)
					}
					header[key] = node
				}
				sig.header = header
			}

			sigBytes, err := base64.RawURLEncoding.DecodeString(rawSig.Signature)
			if err != nil {
				return nil, fmt.Errorf("error decoding signature for signature %d: %v", idx, err)
			}
			sig.signature = sigBytes
			sigs = append(sigs, sig)
		}
	}
	result.signatures = sigs

	return &DagJWS{&result}, nil
}

// Given a JSON string reresenting a JWE in either general or compact serialization this
// will return a DagJWE
func ParseJWE(jsonStr []byte) (*DagJWE, error) {
	var rawJwe struct {
		Protected   *string `json:"protected"`
		Unprotected *string `json:"unprotected"`
		Iv          *string `json:"iv"`
		Aad         *string `json:"aad"`
		Ciphertext  *string `json:"ciphertext"`
		Tag         *string `json:"tag"`
		Recipients  []struct {
			Header       map[string]interface{} `json:"header"`
			EncryptedKey *string                `json:"encrypted_key"`
		} `json:"recipients"`
		Header       map[string]interface{} `json:"header"`
		EncryptedKey *string                `json:"encrypted_key"`
	}

	if err := json.Unmarshal(jsonStr, &rawJwe); err != nil {
		return nil, fmt.Errorf("error parsing JWE json: %v", err)
	}

	if (rawJwe.Header != nil || rawJwe.EncryptedKey != nil) && rawJwe.Recipients != nil {
		return nil, fmt.Errorf("JWE JSON cannot contain 'recipients' and either 'encrypted_key' or 'header'")
	}

	resultJose := DagJOSE{}

	if rawJwe.Ciphertext == nil {
		return nil, fmt.Errorf("JWE has no ciphertext property")
	}
	ciphertextBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("error parsing ciphertext: %v", err)
	}
	resultJose.ciphertext = ciphertextBytes

	if rawJwe.Protected != nil {
		protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Protected)
		if err != nil {
			return nil, fmt.Errorf("error parsing protected: %v", err)
		}
		resultJose.protected = protectedBytes
	}

	var recipients []JWERecipient
	if rawJwe.Header != nil || rawJwe.EncryptedKey != nil {
		recipient := JWERecipient{}
		if rawJwe.EncryptedKey != nil {
			keyBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.EncryptedKey)
			if err != nil {
				return nil, fmt.Errorf("error parsing encrypted_key: %v", err)
			}
			recipient.encrypted_key = keyBytes
		}

		if rawJwe.Header != nil {
			header := make(map[string]ipld.Node)
			for key, v := range rawJwe.Header {
				node, err := goPrimitiveToIpldBasicNode(v)
				if err != nil {
					return nil, fmt.Errorf("error converting header value for key '%s'  of recipient to ipld: %v", key, err)
				}
				header[key] = node
			}
			recipient.header = header
		}
		recipients = append(recipients, recipient)
	} else if rawJwe.Recipients != nil {
		recipients = make([]JWERecipient, 0, len(rawJwe.Recipients))
		for idx, rawRecipient := range rawJwe.Recipients {
			recipient := JWERecipient{}
			if rawRecipient.EncryptedKey != nil {
				keyBytes, err := base64.RawURLEncoding.DecodeString(*rawRecipient.EncryptedKey)
				if err != nil {
					return nil, fmt.Errorf("error parsing encrypted_key for recipient %d: %v", idx, err)
				}
				recipient.encrypted_key = keyBytes
			}

			if rawRecipient.Header != nil {
				header := make(map[string]ipld.Node)
				for key, v := range rawRecipient.Header {
					node, err := goPrimitiveToIpldBasicNode(v)
					if err != nil {
						return nil, fmt.Errorf("error converting header value for key '%s'  of recipient %d to ipld: %v", key, idx, err)
					}
					header[key] = node
				}
				recipient.header = header
			}
			recipients = append(recipients, recipient)
		}
	}
	resultJose.recipients = recipients

	if rawJwe.Unprotected != nil {
		unprotectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Unprotected)
		if err != nil {
			return nil, fmt.Errorf("error parsing unprotected: %v", err)
		}
		resultJose.unprotected = unprotectedBytes
	}

	if rawJwe.Iv != nil {
		ivBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Iv)
		if err != nil {
			return nil, fmt.Errorf("error parsing iv: %v", err)
		}
		resultJose.iv = ivBytes
	}

	if rawJwe.Aad != nil {
		aadBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Aad)
		if err != nil {
			return nil, fmt.Errorf("error parsing aad: %v", err)
		}
		resultJose.aad = aadBytes
	}

	if rawJwe.Tag != nil {
		tagBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.Tag)
		if err != nil {
			return nil, fmt.Errorf("error parsing tag: %v", err)
		}
		resultJose.tag = tagBytes
	}

	return &DagJWE{&resultJose}, nil
}

func (d *DagJWS) asJson() map[string]interface{} {
	jsonJose := make(map[string]interface{})
	jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.dagjose.payload.Bytes())

	if d.dagjose.signatures != nil {
		sigs := make([]map[string]interface{}, 0, len(d.dagjose.signatures))
		for _, sig := range d.dagjose.signatures {
			jsonSig := make(map[string]interface{}, len(d.dagjose.signatures))
			if sig.protected != nil {
				jsonSig["protected"] = base64.RawURLEncoding.EncodeToString(sig.protected)
			}
			if sig.signature != nil {
				jsonSig["signature"] = base64.RawURLEncoding.EncodeToString(sig.signature)
			}
			if sig.header != nil {
				jsonHeader := make(map[string]interface{}, len(sig.header))
				for key, val := range sig.header {
					goVal, err := ipldNodeToGo(val)
					if err != nil {
						panic(fmt.Errorf("GeneralJSONSerialization: error converting %v to go: %v", val, err))
					}
					jsonHeader[key] = goVal
				}
				jsonSig["header"] = jsonHeader
			}
			sigs = append(sigs, jsonSig)
		}
		jsonJose["signatures"] = sigs
	}
	return jsonJose
}

// Return the general json serialization of this JWS
func (d *DagJWS) GeneralJSONSerialization() []byte {
	jsonRep := d.asJson()
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling JWS to json: %v", err))
	}
	return result
}

// Return the flattened json serialization of this JWS
func (d *DagJWS) FlattenedSerialization() ([]byte, error) {
	if len(d.dagjose.signatures) != 1 {
		return nil, fmt.Errorf("Cannot create a flattened serialization for a JWS with more than one signature")
	}
	jsonRep := d.asJson()
	jsonSignature := jsonRep["signatures"].([]map[string]interface{})[0]
	jsonRep["protected"] = jsonSignature["protected"]
	jsonRep["header"] = jsonSignature["header"]
	jsonRep["signature"] = jsonSignature["signature"]
	delete(jsonRep, "signatures")
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling flattened JWS serialization to JSON: %v", err))
	}
	return result, nil
}

func (d *DagJWS) PayloadLink() ipld.Link {
	return cidlink.Link{Cid: *d.dagjose.payload}
}

// Return the general json serialization of this JWE
func (d *DagJWE) GeneralJSONSerialization() []byte {
	jsonRep := d.asJson()
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling JWE to json: %v", err))
	}
	return result
}

// Return the flattened json serialization of this JWE
func (d *DagJWE) FlattenedSerialization() ([]byte, error) {
	jsonRep := d.asJson()
	jsonRecipient := jsonRep["recipients"].([]map[string]interface{})[0]
	jsonRep["header"] = jsonRecipient["header"]
	jsonRep["encrypted_key"] = jsonRecipient["encrypted_key"]
	delete(jsonRep, "recipients")
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling flattened JWE serialization to JSON: %v", err))
	}
	return result, nil
}

func (d *DagJWE) asJson() map[string]interface{} {
	jsonJose := make(map[string]interface{})

	if d.dagjose.protected != nil {
		jsonJose["protected"] = base64.RawURLEncoding.EncodeToString(d.dagjose.protected)
	}
	if d.dagjose.unprotected != nil {
		jsonJose["unprotected"] = base64.RawURLEncoding.EncodeToString(d.dagjose.unprotected)
	}
	if d.dagjose.iv != nil {
		jsonJose["iv"] = base64.RawURLEncoding.EncodeToString(d.dagjose.iv)
	}
	if d.dagjose.aad != nil {
		jsonJose["aad"] = base64.RawURLEncoding.EncodeToString(d.dagjose.aad)
	}
	jsonJose["ciphertext"] = base64.RawURLEncoding.EncodeToString(d.dagjose.ciphertext)
	if d.dagjose.tag != nil {
		jsonJose["tag"] = base64.RawURLEncoding.EncodeToString(d.dagjose.tag)
	}

	if d.dagjose.recipients != nil {
		recipients := make([]map[string]interface{}, 0, len(d.dagjose.recipients))
		for _, r := range d.dagjose.recipients {
			recipientJson := make(map[string]interface{})
			if r.encrypted_key != nil {
				recipientJson["encrypted_key"] = base64.RawURLEncoding.EncodeToString(r.encrypted_key)
			}
			if r.header != nil {
				jsonHeader := make(map[string]interface{}, len(r.header))
				for key, val := range r.header {
					goVal, err := ipldNodeToGo(val)
					if err != nil {
						panic(fmt.Errorf("GeneralJSONSerialization: unable to convert %v from recipient header to go value: %v", val, err))
					}
					jsonHeader[key] = goVal
				}
				recipientJson["header"] = jsonHeader
			}
			recipients = append(recipients, recipientJson)
		}
		jsonJose["recipients"] = recipients
	}
	return jsonJose
}

func goPrimitiveToIpldBasicNode(value interface{}) (ipld.Node, error) {
	switch v := value.(type) {
	case int:
		return basicnode.NewInt(v), nil
	case float32:
		return basicnode.NewFloat(float64(v)), nil
	case float64:
		return basicnode.NewFloat(v), nil
	case bool:
		return basicnode.NewBool(v), nil
	case string:
		return basicnode.NewString(v), nil
	case map[string]interface{}:
		// Note that here we sort the keys before creating the map. This is
		// because ordering of map keys is not defined in Go (or in JSON, which
		// is where this map is coming from in the first place) but order can
		// be meaningful in IPLD, so we specify that the map is in key order
		return fluent.MustBuildMap(
			basicnode.Prototype.Map,
			len(v),
			func(ma fluent.MapAssembler) {
				type kv struct {
					key   string
					value ipld.Node
				}
				kvs := make([]kv, 0)
				for k, v := range v {
					value, err := goPrimitiveToIpldBasicNode(v)
					if err != nil {
						panic(fmt.Errorf("unable to convert primitive value %v to ipld Node: %v", v, err))
					}
					kvs = append(kvs, kv{key: k, value: value})
				}
				sort.SliceStable(kvs, func(i int, j int) bool {
					return kvs[i].key < kvs[j].key
				})
				for _, kv := range kvs {
					ma.AssembleEntry(kv.key).AssignNode(kv.value)
				}
			},
		), nil
	case []interface{}:
		return fluent.MustBuildList(
			basicnode.Prototype.List,
			len(v),
			func(la fluent.ListAssembler) {
				for _, v := range v {
					value, err := goPrimitiveToIpldBasicNode(v)
					if err != nil {
						panic(fmt.Errorf("unable to convert primitive value %v to ipld Node: %v", v, err))
					}
					la.AssembleValue().AssignNode(value)
				}
			},
		), nil
	case nil:
		return ipld.Null, nil
	default:
		return nil, fmt.Errorf("cannot convert %v to an ipld node", v)
	}
}

func ipldNodeToGo(node ipld.Node) (interface{}, error) {
	switch node.ReprKind() {
	case ipld.ReprKind_Bool:
		return node.AsBool()
	case ipld.ReprKind_Bytes:
		return node.AsBytes()
	case ipld.ReprKind_Int:
		return node.AsInt()
	case ipld.ReprKind_Float:
		return node.AsFloat()
	case ipld.ReprKind_String:
		return node.AsString()
	case ipld.ReprKind_Link:
		lnk, err := node.AsLink()
		if err != nil {
			return nil, fmt.Errorf("ipldNodeToGo: error parsing node as link even thought reprkind is link: %v", err)
		}
		return map[string]string{
			"/": lnk.String(),
		}, nil
	case ipld.ReprKind_Map:
		mapIterator := node.MapIterator()
		if mapIterator == nil {
			return nil, fmt.Errorf("ipldNodeToGo: nil MapIterator returned from map node")
		}
		result := make(map[string]interface{})
		for !mapIterator.Done() {
			k, v, err := mapIterator.Next()
			if err != nil {
				return nil, fmt.Errorf("ipldNodeToGo: error whilst iterating over map: %v", err)
			}
			key, err := k.AsString()
			if err != nil {
				return nil, fmt.Errorf("ipldNodeToGo: unable to convert map key to string: %v", err)
			}
			goVal, err := ipldNodeToGo(v)
			if err != nil {
				return nil, fmt.Errorf("ipldNodeToGo: error converting map value to go: %v", err)
			}
			result[key] = goVal
		}
		return result, nil
	case ipld.ReprKind_List:
		listIterator := node.ListIterator()
		if listIterator == nil {
			return nil, fmt.Errorf("ipldNodeToGo: nil listiterator returned from node with list reprkind")
		}
		result := make([]interface{}, 0)
		for !listIterator.Done() {
			_, next, err := listIterator.Next()
			if err != nil {
				return nil, fmt.Errorf("ipldNodeToGo: error iterating over list node: %v", err)
			}
			val, err := ipldNodeToGo(next)
			if err != nil {
				return nil, fmt.Errorf("ipldNodeToGo: error converting list element to go: %v", err)
			}
			result = append(result, val)
		}
		return result, nil
	case ipld.ReprKind_Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("ipldNodeToGo: Unknown ipld node reprkind: %s", node.ReprKind().String())
	}
}

// This exposes a similar interface to the cidlink.LinkBuilder from go-ipld-prime. It's primarily a convenience
// function so you don't have to specify the codec version yourself
func BuildJOSELink(ctx context.Context, linkContext ipld.LinkContext, jose *DagJOSE, storer ipld.Storer) (ipld.Link, error) {
	lb := cidlink.LinkBuilder{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x85, // 0x71 means "dag-jose" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x15, // 0x15 means "sha3-384" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 48,   // sha3-224 hash has a 48-byte sum.
	}}
	return lb.Build(
		ctx,
		linkContext,
		jose,
		storer,
	)
}

// LoadJOSE is a convenience function which wraps ipld.Link.Load. This will provide the dagjose.NodeBuilder
// to the link and attempt to cast the result to a DagJOSE object
func LoadJOSE(lnk ipld.Link, ctx context.Context, linkContext ipld.LinkContext, loader ipld.Loader) (*DagJOSE, error) {
	builder := NewBuilder()
	err := lnk.Load(
		ctx,
		linkContext,
		builder,
		loader,
	)
	if err != nil {
		return nil, err
	}

	n := builder.Build()
	return n.(*DagJOSE), nil
}
