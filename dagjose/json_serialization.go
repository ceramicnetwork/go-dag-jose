package dagjose

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/node/basic"
)

// ParseJWS Given a JSON string representing a JWS in either general or compact
// serialization this will return a DagJWS
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
	_, c, err := cid.CidFromBytes(payloadBytes)
	if err != nil {
		panic(fmt.Errorf("error parsing payload: payload is not a CID"))
	}
	result.payload = &c

	var sigs []jwsSignature
	if rawJws.Signature != nil {
		sig := jwsSignature{}

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
		sigs = make([]jwsSignature, 0, len(rawJws.Signatures))
		for idx, rawSig := range rawJws.Signatures {
			sig := jwsSignature{}
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

// ParseJWE Given a JSON string representing a JWE in either general or compact
// serialization this will return a DagJWE
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

	var recipients []jweRecipient
	if rawJwe.Header != nil || rawJwe.EncryptedKey != nil {
		recipient := jweRecipient{}
		if rawJwe.EncryptedKey != nil {
			keyBytes, err := base64.RawURLEncoding.DecodeString(*rawJwe.EncryptedKey)
			if err != nil {
				return nil, fmt.Errorf("error parsing encrypted_key: %v", err)
			}
			recipient.encryptedKey = keyBytes
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
		recipients = make([]jweRecipient, 0, len(rawJwe.Recipients))
		for idx, rawRecipient := range rawJwe.Recipients {
			recipient := jweRecipient{}
			if rawRecipient.EncryptedKey != nil {
				keyBytes, err := base64.RawURLEncoding.DecodeString(*rawRecipient.EncryptedKey)
				if err != nil {
					return nil, fmt.Errorf("error parsing encrypted_key for recipient %d: %v", idx, err)
				}
				recipient.encryptedKey = keyBytes
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
	jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.payload.Bytes())

	if d.dagJOSE.signatures != nil {
		sigs := make([]map[string]interface{}, 0, len(d.dagJOSE.signatures))
		for _, sig := range d.dagJOSE.signatures {
			jsonSig := make(map[string]interface{}, len(d.dagJOSE.signatures))
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

// GeneralJSONSerialization returns the general json serialization of this JWS
func (d *DagJWS) GeneralJSONSerialization() []byte {
	jsonRep := d.asJson()
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling JWS to json: %v", err))
	}
	return result
}

// FlattenedSerialization returns the flattened json serialization of this JWS
func (d *DagJWS) FlattenedSerialization() ([]byte, error) {
	if len(d.dagJOSE.signatures) != 1 {
		return nil, fmt.Errorf("cannot create a flattened serialization for a JWS with more than one signature")
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

// GeneralJSONSerialization returns the general json serialization of this JWE
func (d *DagJWE) GeneralJSONSerialization() []byte {
	jsonRep := d.asJson()
	result, err := json.Marshal(jsonRep)
	if err != nil {
		panic(fmt.Errorf("error marshaling JWE to json: %v", err))
	}
	return result
}

// FlattenedSerialization return the flattened json serialization of this JWE
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

	if d.dagJOSE.protected != nil {
		jsonJose["protected"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.protected)
	}
	if d.dagJOSE.unprotected != nil {
		jsonJose["unprotected"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.unprotected)
	}
	if d.dagJOSE.iv != nil {
		jsonJose["iv"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.iv)
	}
	if d.dagJOSE.aad != nil {
		jsonJose["aad"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.aad)
	}
	jsonJose["ciphertext"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.ciphertext)
	if d.dagJOSE.tag != nil {
		jsonJose["tag"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.tag)
	}

	if d.dagJOSE.recipients != nil {
		recipients := make([]map[string]interface{}, 0, len(d.dagJOSE.recipients))
		for _, r := range d.dagJOSE.recipients {
			recipientJson := make(map[string]interface{})
			if r.encryptedKey != nil {
				recipientJson["encrypted_key"] = base64.RawURLEncoding.EncodeToString(r.encryptedKey)
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
		return basicnode.NewInt(int64(v)), nil
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
			int64(len(v)),
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
			int64(len(v)),
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
	switch node.Kind() {
	case ipld.Kind_Bool:
		return node.AsBool()
	case ipld.Kind_Bytes:
		return node.AsBytes()
	case ipld.Kind_Int:
		return node.AsInt()
	case ipld.Kind_Float:
		return node.AsFloat()
	case ipld.Kind_String:
		return node.AsString()
	case ipld.Kind_Link:
		lnk, err := node.AsLink()
		if err != nil {
			return nil, fmt.Errorf("ipldNodeToGo: error parsing node as link even thought kind is link: %v", err)
		}
		return map[string]string{
			"/": lnk.String(),
		}, nil
	case ipld.Kind_Map:
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
	case ipld.Kind_List:
		listIterator := node.ListIterator()
		if listIterator == nil {
			return nil, fmt.Errorf("ipldNodeToGo: nil listiterator returned from node with list kind")
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
	case ipld.Kind_Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("ipldNodeToGo: Unknown ipld node kind: %s", node.Kind().String())
	}
}
