package dagjose

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/square/go-jose.v2/json"
	"reflect"

	"github.com/ipfs/go-cid"
	ipldJson "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multibase"
)

func unflattenJWE(n datamodel.Node) (datamodel.Node, error) {
	if recipients, err := lookupIgnoreAbsent("recipients", n); err != nil {
		return nil, err
	} else if recipients != nil {
		// If `recipients` is present, this must be a "general" JWE and no changes are needed but make sure that
		// `header` and/or `encrypted_key` are not also present since that would be a violation of the spec.
		if encryptedKey, err := lookupIgnoreNoSuchField("encrypted_key", n); err != nil {
			return nil, err
		} else if encryptedKey != nil {
			return nil, errors.New("invalid JWE serialization")
		}
		if header, err := lookupIgnoreNoSuchField("header", n); err != nil {
			return nil, err
		} else if header != nil {
			return nil, errors.New("invalid JWE serialization")
		}
		return n, nil
	} else
	// If `recipients` is absent, this must be a "flattened" JWE.
	if ciphertext, err := n.LookupByString("ciphertext"); err != nil {
		// `ciphertext` is mandatory so if any error occurs, return from here
		return nil, err
	} else if ciphertextString, err := ciphertext.AsString(); err != nil {
		return nil, err
	} else {
		recipients := make([]map[string]interface{}, 1)
		recipients[0] = make(map[string]interface{}, 0) // all recipient fields are optional
		jwe := map[string]interface{}{
			"ciphertext": ciphertextString,
		}
		if aad, err := lookupIgnoreAbsent("aad", n); err != nil {
			return nil, err
		} else if aad != nil {
			if aadString, err := aad.AsString(); err != nil {
				return nil, err
			} else {
				jwe["aad"] = aadString
			}
		}
		if encryptedKey, err := lookupIgnoreAbsent("encrypted_key", n); err != nil {
			return nil, err
		} else if encryptedKey != nil {
			if encryptedKeyString, err := encryptedKey.AsString(); err != nil {
				return nil, err
			} else {
				recipients[0]["encrypted_key"] = encryptedKeyString
			}
		}
		if header, err := lookupIgnoreAbsent("header", n); err != nil {
			return nil, err
		} else if header != nil {
			if headerMap, err := nodeToMap(header); err != nil {
				return nil, err
			} else {
				recipients[0]["header"] = headerMap
			}
		}
		if iv, err := lookupIgnoreAbsent("iv", n); err != nil {
			return nil, err
		} else if iv != nil {
			if ivString, err := iv.AsString(); err != nil {
				return nil, err
			} else {
				jwe["iv"] = ivString
			}
		}
		if link, err := lookupIgnoreAbsent("link", n); err != nil {
			return nil, err
		} else if link != nil {
			if linkString, err := link.AsString(); err != nil {
				return nil, err
			} else {
				jwe["link"] = map[string]string{
					"/": linkString,
				}
			}
		}
		if protected, err := lookupIgnoreAbsent("protected", n); err != nil {
			return nil, err
		} else if protected != nil {
			if protectedString, err := protected.AsString(); err != nil {
				return nil, err
			} else {
				jwe["protected"] = protectedString
			}
		}
		if tag, err := lookupIgnoreAbsent("tag", n); err != nil {
			return nil, err
		} else if tag != nil {
			if tagString, err := tag.AsString(); err != nil {
				return nil, err
			} else {
				jwe["tag"] = tagString
			}
		}
		if unprotected, err := lookupIgnoreAbsent("unprotected", n); err != nil {
			return nil, err
		} else if unprotected != nil {
			if unprotectedMap, err := nodeToMap(unprotected); err != nil {
				return nil, err
			} else {
				jwe["unprotected"] = unprotectedMap
			}
		}
		// Only add `recipients` to the JWE if one or more fields were present in the map
		if len(recipients[0]) > 0 {
			jwe["recipients"] = recipients
		}
		return mapToNode(jwe)
	}
}

func unflattenJWS(n datamodel.Node) (datamodel.Node, error) {
	if signatures, err := lookupIgnoreAbsent("signatures", n); err != nil {
		return nil, err
	} else if signatures != nil {
		// If `signatures` is present, this must be a "general" JWS and no changes are needed but make sure that
		// `header`, `protected`, and/or `signature` are not also present since that would be a violation of the spec.
		if header, err := lookupIgnoreNoSuchField("header", n); err != nil {
			return nil, err
		} else if header != nil {
			return nil, errors.New("invalid JWS serialization")
		}
		if protected, err := lookupIgnoreNoSuchField("protected", n); err != nil {
			return nil, err
		} else if protected != nil {
			return nil, errors.New("invalid JWS serialization")
		}
		if signature, err := lookupIgnoreNoSuchField("signature", n); err != nil {
			return nil, err
		} else if signature != nil {
			return nil, errors.New("invalid JWS serialization")
		}
		return n, nil
	} else
	// If `signatures` is absent, this must be a "flattened" JWS.
	if payload, err := n.LookupByString("payload"); err != nil {
		// `payload` is mandatory so if any error occurs, return from here
		return nil, err
	} else if payloadString, err := payload.AsString(); err != nil {
		return nil, err
	} else if _, err := cid.Decode(string(multibase.Base64url) + payloadString); err != nil {
		return nil, errors.New(fmt.Sprintf("payload is not a valid CID: %v", err))
	} else if payloadString, err := payload.AsString(); err != nil {
		return nil, err
	} else {
		signatures := make([]map[string]interface{}, 1)
		signatures[0] = make(map[string]interface{}, 1) // at least `signature` must be present
		jws := map[string]interface{}{
			"payload":    payloadString,
			"signatures": signatures,
		}
		if header, err := lookupIgnoreAbsent("header", n); err != nil {
			return nil, err
		} else if header != nil {
			if headerMap, err := nodeToMap(header); err != nil {
				return nil, err
			} else {
				signatures[0]["header"] = headerMap
			}
		}
		if protected, err := lookupIgnoreAbsent("protected", n); err != nil {
			return nil, err
		} else if protected != nil {
			if protectedString, err := protected.AsString(); err != nil {
				return nil, err
			} else {
				signatures[0]["protected"] = protectedString
			}
		}
		if signature, err := lookupIgnoreAbsent("signature", n); err != nil {
			return nil, err
		} else if signature != nil {
			if signatureString, err := signature.AsString(); err != nil {
				return nil, err
			} else {
				signatures[0]["signature"] = signatureString
			}
		}
		return mapToNode(jws)
	}
}

func isJWS(n datamodel.Node) (bool, error) {
	if payload, err := lookupIgnoreNoSuchField("payload", n); err != nil {
		return false, err
	} else {
		return payload != nil, nil
	}
}

func isJWE(n datamodel.Node) (bool, error) {
	if ciphertext, err := lookupIgnoreNoSuchField("ciphertext", n); err != nil {
		return false, err
	} else {
		return ciphertext != nil, nil
	}
}

func mapToNode(m map[string]interface{}) (datamodel.Node, error) {
	if jsonBytes, err := json.Marshal(m); err != nil {
		return nil, err
	} else {
		na := basicnode.Prototype.Any.NewBuilder()
		if err := ipldJson.Decode(na, bytes.NewReader(jsonBytes)); err != nil {
			return nil, err
		} else {
			return na.Build(), nil
		}
	}
}

func nodeToMap(n datamodel.Node) (map[string]interface{}, error) {
	jsonBytes := bytes.NewBuffer([]byte{})
	if err := (ipldJson.EncodeOptions{
		EncodeLinks: false,
		EncodeBytes: false,
	}.Encode(n, jsonBytes)); err != nil {
		return nil, err
	} else {
		m := map[string]interface{}{}
		if err := json.Unmarshal(jsonBytes.Bytes(), &m); err != nil {
			return nil, err
		} else {
			return sanitizeMap(m), nil
		}
	}
}

// Remove all `nil` values from the top-level structure or from within nested maps or slices
func sanitizeMap(m map[string]interface{}) map[string]interface{} {
	for key, value := range m {
		if value == nil {
			delete(m, key)
		} else if reflect.ValueOf(value).Kind() == reflect.Slice {
			for idx, entry := range value.([]interface{}) {
				if reflect.ValueOf(entry).Kind() == reflect.Map {
					m[key].([]interface{})[idx] = sanitizeMap(m[key].([]interface{})[idx].(map[string]interface{}))
				}
			}
		} else if reflect.ValueOf(value).Kind() == reflect.Map {
			m[key] = sanitizeMap(value.(map[string]interface{}))
		}
	}
	return m
}

func lookupIgnoreAbsent(key string, n datamodel.Node) (datamodel.Node, error) {
	if value, err := n.LookupByString(key); err != nil {
		if _, notFoundErr := err.(datamodel.ErrNotExists); !notFoundErr {
			return nil, err
		}
		return nil, nil
	} else {
		return value, nil
	}
}

func lookupIgnoreNoSuchField(key string, n datamodel.Node) (datamodel.Node, error) {
	value, err := lookupIgnoreAbsent(key, n)
	if err != nil {
		if _, noSuchFieldErr := err.(schema.ErrNoSuchField); !noSuchFieldErr {
			return nil, err
		}
	}
	return value, nil
}
