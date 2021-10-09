package dagjose

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ipld/go-ipld-prime/fluent"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/pkg/errors"
)

// ParseJWS Given a JSON string representing a JWS in either general or compact serialization this
// will return a DAGJWS
func ParseJWS(jsonBytes []byte) (*DAGJWS, error) {
	var rawJWS struct {
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
	if err := json.Unmarshal(jsonBytes, &rawJWS); err != nil {
		return nil, errors.Wrap(err, "error parsing jws json")
	}
	result := DAGJOSE{}

	if rawJWS.Payload == nil {
		return nil, errors.New("JWS has no payload property")
	}

	if rawJWS.Signature != nil && rawJWS.Signatures != nil {
		return nil, errors.New("JWS JSON cannot contain both a 'signature' and a 'signatures' key")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(*rawJWS.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing payload")
	}
	_, id, err := cid.CidFromBytes(payloadBytes)
	if err != nil {
		return nil, errors.New("error parsing payload: payload is not a CID")
	}
	result.payload = &id

	var sigs []jwsSignature
	if rawJWS.Signature != nil {
		sig := jwsSignature{}

		sigBytes, err := base64.RawURLEncoding.DecodeString(*rawJWS.Signature)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding signature")
		}
		sig.signature = sigBytes

		if rawJWS.Protected != nil {
			protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJWS.Protected)
			if err != nil {
				return nil, errors.Wrap(err, "error parsing signature")
			}
			sig.protected = protectedBytes
		}

		if rawJWS.Header != nil {
			header := make(map[string]ipld.Node)
			for key, v := range rawJWS.Header {
				node, err := goPrimitiveToIPLDBasicNode(v)
				if err != nil {
					return nil, errors.Wrapf(err, "error converting header value for key '%s'  of to ipld", key)
				}
				header[key] = node
			}
			sig.header = header
		}
		sigs = append(sigs, sig)
	} else if rawJWS.Signatures != nil {
		sigs = make([]jwsSignature, 0, len(rawJWS.Signatures))
		for idx, rawSig := range rawJWS.Signatures {
			sig := jwsSignature{}
			if rawSig.Protected != nil {
				protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawSig.Protected)
				if err != nil {
					return nil, errors.Wrapf(err, "error parsing signatures[%d]['protected']", idx)
				}
				sig.protected = protectedBytes
			}

			if rawSig.Header != nil {
				header := make(map[string]ipld.Node)
				for key, v := range rawSig.Header {
					node, err := goPrimitiveToIPLDBasicNode(v)
					if err != nil {
						return nil, errors.Wrapf(err, "error converting header value for key '%s'  of sign %d to ipld", key, idx)
					}
					header[key] = node
				}
				sig.header = header
			}

			sigBytes, err := base64.RawURLEncoding.DecodeString(rawSig.Signature)
			if err != nil {
				return nil, errors.Wrapf(err, "error decoding signature for signature %d", idx)
			}
			sig.signature = sigBytes
			sigs = append(sigs, sig)
		}
	}
	result.signatures = sigs

	return &DAGJWS{dagJOSE: &result}, nil
}

// ParseJWE Given a JSON string representing a JWE in either general or compact serialization this
// will return a DAGJWE
func ParseJWE(jsonBytes []byte) (*DAGJWE, error) {
	var rawJWE struct {
		Protected   *string `json:"protected"`
		Unprotected *string `json:"unprotected"`
		IV          *string `json:"iv"`
		AAD         *string `json:"aad"`
		Ciphertext  *string `json:"ciphertext"`
		Tag         *string `json:"tag"`
		Recipients  []struct {
			Header       map[string]interface{} `json:"header"`
			EncryptedKey *string                `json:"encrypted_key"`
		} `json:"recipients"`
		Header       map[string]interface{} `json:"header"`
		EncryptedKey *string                `json:"encrypted_key"`
	}

	if err := json.Unmarshal(jsonBytes, &rawJWE); err != nil {
		return nil, errors.Wrap(err, "error parsing JWE json")
	}

	if (rawJWE.Header != nil || rawJWE.EncryptedKey != nil) && rawJWE.Recipients != nil {
		return nil, errors.New("JWE JSON cannot contain 'recipients' and either 'encrypted_key' or 'header'")
	}

	resultJOSE := DAGJOSE{}
	if rawJWE.Ciphertext == nil {
		return nil, fmt.Errorf("JWE has no ciphertext property")
	}
	ciphertextBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.Ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing ciphertext")
	}
	resultJOSE.ciphertext = ciphertextBytes

	if rawJWE.Protected != nil {
		protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.Protected)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing protected")
		}
		resultJOSE.protected = protectedBytes
	}

	var recipients []jweRecipient
	if rawJWE.Header != nil || rawJWE.EncryptedKey != nil {
		recipient := jweRecipient{}
		if rawJWE.EncryptedKey != nil {
			keyBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.EncryptedKey)
			if err != nil {
				return nil, errors.Wrap(err, "error parsing encrypted_key")
			}
			recipient.encryptedKey = keyBytes
		}

		if rawJWE.Header != nil {
			header := make(map[string]ipld.Node)
			for key, v := range rawJWE.Header {
				node, err := goPrimitiveToIPLDBasicNode(v)
				if err != nil {
					return nil, errors.Wrapf(err, "error converting header value for key '%s'  of recipient to ipld", key)
				}
				header[key] = node
			}
			recipient.header = header
		}
		recipients = append(recipients, recipient)
	} else if rawJWE.Recipients != nil {
		recipients = make([]jweRecipient, 0, len(rawJWE.Recipients))
		for idx, rawRecipient := range rawJWE.Recipients {
			recipient := jweRecipient{}
			if rawRecipient.EncryptedKey != nil {
				keyBytes, err := base64.RawURLEncoding.DecodeString(*rawRecipient.EncryptedKey)
				if err != nil {
					return nil, errors.Wrapf(err, "error parsing encrypted_key for recipient %d", idx)
				}
				recipient.encryptedKey = keyBytes
			}

			if rawRecipient.Header != nil {
				header := make(map[string]ipld.Node)
				for key, v := range rawRecipient.Header {
					node, err := goPrimitiveToIPLDBasicNode(v)
					if err != nil {
						return nil, errors.Wrapf(err, "error converting header value for key '%s'  of recipient %d to ipld", key, idx)
					}
					header[key] = node
				}
				recipient.header = header
			}
			recipients = append(recipients, recipient)
		}
	}
	resultJOSE.recipients = recipients

	if rawJWE.Unprotected != nil {
		unprotectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.Unprotected)
		if err != nil {
			return nil, fmt.Errorf("error parsing unprotected: %v", err)
		}
		resultJOSE.unprotected = unprotectedBytes
	}

	if rawJWE.IV != nil {
		ivBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.IV)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing iv")
		}
		resultJOSE.iv = ivBytes
	}

	if rawJWE.AAD != nil {
		aadBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.AAD)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing aad")
		}
		resultJOSE.aad = aadBytes
	}

	if rawJWE.Tag != nil {
		tagBytes, err := base64.RawURLEncoding.DecodeString(*rawJWE.Tag)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing tag")
		}
		resultJOSE.tag = tagBytes
	}

	return &DAGJWE{dagjose: &resultJOSE}, nil
}

func (d *DAGJWS) asJSON() (map[string]interface{}, error) {
	jsonJOSE := make(map[string]interface{})
	jsonJOSE["payload"] = base64.RawURLEncoding.EncodeToString(d.dagJOSE.payload.Bytes())

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
						return nil, errors.Wrapf(err, "GeneralJSONSerialization: error converting %v to go", val)
					}
					jsonHeader[key] = goVal
				}
				jsonSig["header"] = jsonHeader
			}
			sigs = append(sigs, jsonSig)
		}
		jsonJOSE["signatures"] = sigs
	}
	return jsonJOSE, nil
}

// GeneralJSONSerialization Return the general json serialization of this JWS
func (d *DAGJWS) GeneralJSONSerialization() ([]byte, error) {
	jsonRep, err := d.asJSON()
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(jsonRep)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling JWS to json")
	}
	return result, nil
}

// FlattenedSerialization Return the flattened json serialization of this JWS
func (d *DAGJWS) FlattenedSerialization() ([]byte, error) {
	if len(d.dagJOSE.signatures) != 1 {
		return nil, errors.New("cannot create a flattened serialization for a JWS with more than one signature")
	}
	jsonRep, err := d.asJSON()
	if err != nil {
		return nil, err
	}
	jsonSignature := jsonRep["signatures"].([]map[string]interface{})[0]
	jsonRep["protected"] = jsonSignature["protected"]
	jsonRep["header"] = jsonSignature["header"]
	jsonRep["signature"] = jsonSignature["signature"]
	delete(jsonRep, "signatures")
	result, err := json.Marshal(jsonRep)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshaling flattened JWS serialization to JSON")
	}
	return result, nil
}

// GeneralJSONSerialization Return the general json serialization of this JWE
func (d *DAGJWE) GeneralJSONSerialization() ([]byte, error) {
	jsonRep, err := d.asJSON()
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(jsonRep)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling JWE to json")
	}
	return result, nil
}

// FlattenedSerialization Return the flattened json serialization of this JWE
func (d *DAGJWE) FlattenedSerialization() ([]byte, error) {
	jsonRep, err := d.asJSON()
	if err != nil {
		return nil, err
	}
	jsonRecipient := jsonRep["recipients"].([]map[string]interface{})[0]
	jsonRep["header"] = jsonRecipient["header"]
	jsonRep["encrypted_key"] = jsonRecipient["encrypted_key"]
	delete(jsonRep, "recipients")
	result, err := json.Marshal(jsonRep)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling flattened JWE serialization to JSON")
	}
	return result, nil
}

func (d *DAGJWE) asJSON() (map[string]interface{}, error) {
	jsonJOSE := make(map[string]interface{})

	if d.dagjose.protected != nil {
		jsonJOSE["protected"] = base64.RawURLEncoding.EncodeToString(d.dagjose.protected)
	}
	if d.dagjose.unprotected != nil {
		jsonJOSE["unprotected"] = base64.RawURLEncoding.EncodeToString(d.dagjose.unprotected)
	}
	if d.dagjose.iv != nil {
		jsonJOSE["iv"] = base64.RawURLEncoding.EncodeToString(d.dagjose.iv)
	}
	if d.dagjose.aad != nil {
		jsonJOSE["aad"] = base64.RawURLEncoding.EncodeToString(d.dagjose.aad)
	}
	jsonJOSE["ciphertext"] = base64.RawURLEncoding.EncodeToString(d.dagjose.ciphertext)
	if d.dagjose.tag != nil {
		jsonJOSE["tag"] = base64.RawURLEncoding.EncodeToString(d.dagjose.tag)
	}

	if d.dagjose.recipients != nil {
		recipients := make([]map[string]interface{}, 0, len(d.dagjose.recipients))
		for _, r := range d.dagjose.recipients {
			recipientJSON := make(map[string]interface{})
			if r.encryptedKey != nil {
				recipientJSON["encrypted_key"] = base64.RawURLEncoding.EncodeToString(r.encryptedKey)
			}
			if r.header != nil {
				jsonHeader := make(map[string]interface{}, len(r.header))
				for key, val := range r.header {
					goVal, err := ipldNodeToGo(val)
					if err != nil {
						return nil, errors.Wrapf(err, "GeneralJSONSerialization: unable to convert %v from recipient header to go value", val)
					}
					jsonHeader[key] = goVal
				}
				recipientJSON["header"] = jsonHeader
			}
			recipients = append(recipients, recipientJSON)
		}
		jsonJOSE["recipients"] = recipients
	}
	return jsonJOSE, nil
}

func goPrimitiveToIPLDBasicNode(value interface{}) (ipld.Node, error) {
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
					value, err := goPrimitiveToIPLDBasicNode(v)
					if err != nil {
						// TODO it's unclear the correct behavior here, a nil return may be more beneficial
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
					value, err := goPrimitiveToIPLDBasicNode(v)
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
				return nil, errors.Wrap(err, "ipldNodeToGo: error whilst iterating over map")
			}
			key, err := k.AsString()
			if err != nil {
				return nil, errors.Wrap(err, "ipldNodeToGo: unable to convert map key to string")
			}
			goVal, err := ipldNodeToGo(v)
			if err != nil {
				return nil, errors.Wrap(err, "ipldNodeToGo: error converting map value to go")
			}
			result[key] = goVal
		}
		return result, nil
	case ipld.Kind_List:
		listIterator := node.ListIterator()
		if listIterator == nil {
			return nil, errors.New("ipldNodeToGo: nil list iterator returned from node with list kind")
		}
		var result []interface{}
		for !listIterator.Done() {
			_, next, err := listIterator.Next()
			if err != nil {
				return nil, errors.Wrap(err, "ipldNodeToGo: error iterating over list node")
			}
			val, err := ipldNodeToGo(next)
			if err != nil {
				return nil, errors.Wrap(err, "ipldNodeToGo: error converting list element to go")
			}
			result = append(result, val)
		}
		return result, nil
	case ipld.Kind_Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("ipldNodeToGo: Unknown IPLD node kind: %s", node.Kind().String())
	}
}
