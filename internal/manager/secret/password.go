package secret

import "encoding/json"

// Password is a plaintext secret value.
type Password struct {
	Username string `json:"u"`
	Password string `json:"p"`
}

// NewPassword creates a new Password secret.
func NewPassword(username, password string) Secret {
	v := Password{
		Username: username,
		Password: password,
	}
	return Secret{
		secret: &v,
	}
}

func (p Password) typeMarker() byte {
	return 'p'
}

func (p Password) marshal() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		return nil
	}

	return b
}

func (p *Password) unmarshal(data []byte) error {
	return json.Unmarshal(data, p)
}
