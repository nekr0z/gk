package secret

import (
	"encoding/json"
)

// Card is a secret value representing a payment card.
type Card struct {
	Number string `json:"n"`
	Expiry string `json:"e"`
	CVV    string `json:"c"`
}

// NewCard creates a new Card secret.
func NewCard(number, expiry, cvv string) Secret {
	return Secret{
		secret: &Card{
			Number: number,
			Expiry: expiry,
			CVV:    cvv,
		},
	}
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
