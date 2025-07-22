package secret

import (
	"encoding/json"
	"strings"
)

// Card is a secret value representing a payment card.
type Card struct {
	Number   string `json:"n"`
	Expiry   string `json:"e"`
	CVV      string `json:"c"`
	Username string `json:"u"`
}

// NewCard creates a new Card secret.
func NewCard(number, expiry, cvv, name string) Secret {
	return Secret{
		secret: &Card{
			Number:   number,
			Expiry:   expiry,
			CVV:      cvv,
			Username: name,
		},
		metadata: make(map[string]string),
	}
}

// String returns the string representation of the Binary.
func (c Card) String() string {
	var sb strings.Builder
	sb.WriteString("Card Number: ")
	sb.WriteString(c.Number)
	sb.WriteString("\n")
	sb.WriteString("Expiry Date: ")
	sb.WriteString(c.Expiry)
	sb.WriteString("\n")
	sb.WriteString("CVV: ")
	sb.WriteString(c.CVV)

	if c.Username != "" {
		sb.WriteString("\n")
		sb.WriteString("Username: ")
		sb.WriteString(c.Username)
	}

	sb.WriteString("\n")

	return sb.String()
}

func (c Card) typeMarker() byte {
	return 'c'
}

func (c Card) marshal() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		return nil
	}

	return b
}

func (c *Card) unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}
