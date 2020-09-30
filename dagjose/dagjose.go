package dagjose

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	//ipld "github.com/ipld/go-ipld-prime"
)

type JOSESignature struct {
	protected []byte
	header    map[string]string
	signature []byte
}

type JWERecipient struct {
	//header        map[string]ipld.Node
	header        map[string]string
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
	var rawJws struct {
		Payload   string `json:"payload"`
		Protected string `json:"protected"`
		Signature string `json:"signature"`
	}
	err := json.Unmarshal([]byte(jsonSerialization), &rawJws)
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

func (d *DagJOSE) GeneralJSONSerialization() string {
	jsonJose := make(map[string]interface{})
	jsonJose["payload"] = base64.RawURLEncoding.EncodeToString(d.payload)
	sigs := make([]map[string]string, 0)
	if d.signatures != nil {
		for _, sig := range d.signatures {
			jsonSig := make(map[string]string, len(d.signatures))
			if sig.protected != nil {
				jsonSig["protected"] = base64.RawURLEncoding.EncodeToString(sig.protected)
			}
			if sig.signature != nil {
				jsonSig["signature"] = base64.RawURLEncoding.EncodeToString(sig.signature)
			}
			if sig.header != nil {
				headerJson, err := json.Marshal(sig.header)
				if err != nil {
					panic(fmt.Errorf("Error marshaling protected header to json: %v", err))
				}
				jsonSig["header"] = base64.RawURLEncoding.EncodeToString(headerJson)
			}
			sigs = append(sigs, jsonSig)
		}
	}

	if d.signatures != nil {
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
		recipients := make([]map[string]string, len(d.recipients))
		for _, r := range d.recipients {
			recipientJson := make(map[string]string)
			if r.encrypted_key != nil {
				recipientJson["encrypted_key"] = base64.RawURLEncoding.EncodeToString(r.encrypted_key)
			}
			if r.header != nil {
				headerJson, err := json.Marshal(r.header)
				if err != nil {
					panic(fmt.Errorf("Error marshaling protected header to json: %v", err))
				}
				recipientJson["header"] = base64.RawURLEncoding.EncodeToString(headerJson)
			}
			recipients = append(recipients, recipientJson)
		}
		jsonJose["recipients"] = recipients
	}
	encoded, err := json.Marshal(jsonJose)
	if err != nil {
		panic(fmt.Errorf("error marshaling jose serialization to json: %v", err))
	}
	return string(encoded)
}
