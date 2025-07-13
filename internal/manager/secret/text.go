package secret

// Text is a plaintext secret value.
type Text string

// NewText creates a new Text secret.
func NewText(value string) Secret {
	v := Text(value)
	return Secret{
		secret: &v,
	}
}

// String returns the string representation of the Text.
func (t Text) String() string {
	return string(t)
}

func (t Text) typeMarker() byte {
	return 't'
}

func (t Text) marshal() []byte {
	return []byte(t)
}

func (t *Text) unmarshal(data []byte) error {
	*t = Text(data)
	return nil
}
