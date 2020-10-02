package dagjose

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/fluent"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

type JOSESignature struct {
	protected []byte
	header    map[string]ipld.Node
	signature []byte
}

type JWERecipient struct {
	header        map[string]ipld.Node
	encrypted_key []byte
}

type DagJOSE struct {
	// JWS top level keys
	payload    []byte
	signatures []JOSESignature
	// JWE top level keys
	protected   []byte
	unprotected []byte
	iv          []byte
	aad         []byte
	ciphertext  []byte
	tag         []byte
	recipients  []JWERecipient
}

func NewDagJWS(jsonSerialization string) (*DagJOSE, error) {
	result, err := parseGeneralSerialization(jsonSerialization)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	var rawJws struct {
		Payload   string `json:"payload"`
		Protected string `json:"protected"`
		Signature string `json:"signature"`
	}
	err = json.Unmarshal([]byte(jsonSerialization), &rawJws)
	if err != nil {
		return nil, fmt.Errorf("error deserializing json: %v", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(rawJws.Payload)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload bytes: %v", err)
	}
	protectedBytes, err := base64.RawURLEncoding.DecodeString(rawJws.Protected)
	if err != nil {
		return nil, fmt.Errorf("error decoding protected bytes: %v", err)
	}
	signatureBytes, err := base64.RawURLEncoding.DecodeString(rawJws.Signature)
	if err != nil {
		return nil, fmt.Errorf("error decoding signature bytes: %v", err)
	}
	return &DagJOSE{
		payload: payloadBytes,
		signatures: []JOSESignature{
			{
				protected: protectedBytes,
				signature: signatureBytes,
				header:    nil,
			},
		},
	}, nil
}

func parseGeneralSerialization(jsonStr string) (*DagJOSE, error) {
	var rawJose struct {
		Payload    *string `json:"payload"`
		Protected  *string `json:"protected"`
		Signatures []struct {
			Protected *string                `json:"protected"`
			Signature string                 `json:"signature"`
			Header    map[string]interface{} `json:"header"`
		} `json:"signatures"`
		Unprotected *string `json:"unprotected"`
		Iv          *string `json:"iv"`
		Aad         *string `json:"aad"`
		Ciphertext  *string `json:"ciphertext"`
		Tag         *string `json:"tag"`
		Recipients  []struct {
			Header       map[string]interface{} `json:"header"`
			EncryptedKey *string                `json:"encrypted_key"`
		} `json:"recipients"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &rawJose); err != nil {
		return nil, fmt.Errorf("error parsing for general serialization: %v", err)
	}
	result := DagJOSE{}
	if rawJose.Payload != nil {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Payload)
		if err != nil {
			return nil, fmt.Errorf("error parsing payload: %v", err)
		}
		result.payload = payloadBytes
	}

	if rawJose.Protected != nil {
		protectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Protected)
		if err != nil {
			return nil, fmt.Errorf("error parsing protected: %v", err)
		}
		result.protected = protectedBytes
	}

	if rawJose.Signatures != nil {
		sigs := make([]JOSESignature, 0, len(rawJose.Signatures))
		for idx, rawSig := range rawJose.Signatures {
			sig := JOSESignature{}
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
		result.signatures = sigs
	}

	if rawJose.Recipients != nil {
		recipients := make([]JWERecipient, 0, len(rawJose.Recipients))
		for idx, rawRecipient := range rawJose.Recipients {
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
		result.recipients = recipients
	}

	if rawJose.Unprotected != nil {
		unprotectedBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Unprotected)
		if err != nil {
			return nil, fmt.Errorf("error parsing unprotected: %v", err)
		}
		result.unprotected = unprotectedBytes
	}

	if rawJose.Iv != nil {
		ivBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Iv)
		if err != nil {
			return nil, fmt.Errorf("error parsing iv: %v", err)
		}
		result.iv = ivBytes
	}

	if rawJose.Aad != nil {
		aadBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Aad)
		if err != nil {
			return nil, fmt.Errorf("error parsing aad: %v", err)
		}
		result.aad = aadBytes
	}

	if rawJose.Ciphertext != nil {
		ciphertextBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Ciphertext)
		if err != nil {
			return nil, fmt.Errorf("error parsing ciphertext: %v", err)
		}
		result.ciphertext = ciphertextBytes
	}

	if rawJose.Tag != nil {
		tagBytes, err := base64.RawURLEncoding.DecodeString(*rawJose.Tag)
		if err != nil {
			return nil, fmt.Errorf("error parsing tag: %v", err)
		}
		result.tag = tagBytes
	}

	return &result, nil
}

func (d *DagJOSE) GeneralJSONSerialization() string {
	jsonJose := make(map[string]interface{})
	if d.payload != nil {
		jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.payload)
	}

	if d.signatures != nil {
		sigs := make([]map[string]interface{}, 0, len(d.signatures))
		for _, sig := range d.signatures {
			jsonSig := make(map[string]interface{}, len(d.signatures))
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
	if d.protected != nil {
		jsonJose["protected"] = base64.RawURLEncoding.EncodeToString(d.protected)
	}
	if d.unprotected != nil {
		jsonJose["unprotected"] = base64.RawURLEncoding.EncodeToString(d.unprotected)
	}
	if d.iv != nil {
		jsonJose["iv"] = base64.RawURLEncoding.EncodeToString(d.iv)
	}
	if d.aad != nil {
		jsonJose["aad"] = base64.RawURLEncoding.EncodeToString(d.aad)
	}
	if d.ciphertext != nil {
		jsonJose["ciphertext"] = base64.RawURLEncoding.EncodeToString(d.ciphertext)
	}
	if d.tag != nil {
		jsonJose["tag"] = base64.RawURLEncoding.EncodeToString(d.tag)
	}

	if d.recipients != nil {
		recipients := make([]map[string]interface{}, 0, len(d.recipients))
		for _, r := range d.recipients {
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
	encoded, err := json.Marshal(jsonJose)
	if err != nil {
		panic(fmt.Errorf("GeneralJSONSerialization: error marshaling jose serialization to json: %v", err))
	}
	return string(encoded)
}

func (d *DagJOSE) FlattenedSerialization() (string, error) {
	return "", nil
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
