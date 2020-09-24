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
	for _, sig := range d.signatures {
		jsonSig := make(map[string]string)
		if sig.protected != nil {
			jsonSig["protected"] = base64.RawURLEncoding.EncodeToString(sig.protected)
		}
		jsonSig["signature"] = base64.RawURLEncoding.EncodeToString(sig.signature)
		sigs = append(sigs, jsonSig)
	}
	jsonJose["signatures"] = sigs
	encoded, err := json.Marshal(jsonJose)
	if err != nil {
		panic("impossible")
	}
	return string(encoded)
}
