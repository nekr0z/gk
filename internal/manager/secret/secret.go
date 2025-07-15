// Package secret provides an abstraction over secrets.
package secret

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Secret is a secret.
type Secret struct {
	secret   secret
	metadata map[string]string
}

type secret interface {
	typeMarker() byte
	marshal() []byte
	unmarshal([]byte) error

	fmt.Stringer
}

type secretJSON struct {
	Type     byte              `json:"t"`
	Data     []byte            `json:"d"`
	Metadata map[string]string `json:"m"`
}

// Marshal returns a marshaled Secret.
func (s Secret) Marshal() []byte {
	j := secretJSON{
		Type:     s.secret.typeMarker(),
		Data:     s.secret.marshal(),
		Metadata: s.metadata,
	}

	b, err := json.Marshal(j)
	if err != nil {
		return nil
	}

	return b
}

// Unmarshal unmarshals a Secret.
func Unmarshal(data []byte) (Secret, error) {
	var j secretJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		return Secret{}, err
	}

	var s Secret

	switch j.Type {
	case Text("").typeMarker():
		t := Text("")
		s.secret = &t
	case Binary(nil).typeMarker():
		b := Binary(nil)
		s.secret = &b
	case Password{}.typeMarker():
		p := Password{}
		s.secret = &p
	case Card{}.typeMarker():
		c := Card{}
		s.secret = &c
	default:
		return Secret{}, fmt.Errorf("unknown secret type")
	}

	err = s.secret.unmarshal(j.Data)
	if err != nil {
		return Secret{}, err
	}

	s.metadata = j.Metadata

	return s, nil
}

// Value returns the value of the secret.
func (s Secret) Value() secret {
	return s.secret
}

// Metadata returns the metadata of the secret.
func (s Secret) Metadata() map[string]string {
	return s.metadata
}

// SetMetadata sets the metadata of the secret.
func (s *Secret) SetMetadata(metadata map[string]string) {
	s.metadata = metadata
}

// SetMetadataValue sets a particular metadata value of the secret.
func (s *Secret) SetMetadataValue(key, value string) {
	s.metadata[key] = value
}

// GetMetadataValue gets a particular metadata value of the secret.
func (s *Secret) GetMetadataValue(key string) (string, bool) {
	v, ok := s.metadata[key]
	return v, ok
}

// String returns the string representation of the secret.
func (s Secret) String() string {
	var sb strings.Builder
	sb.WriteString(s.secret.String())
	sb.WriteString("\n")
	for k, v := range s.metadata {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	return sb.String()
}
