package secret

// Binary is a secret value that contains binary data.
type Binary []byte

// NewBinary creates a new Binary secret.
func NewBinary(value []byte) Secret {
	v := Binary(value)
	return Secret{
		secret:   &v,
		metadata: make(map[string]string),
	}
}

// Bytes returns the bytes content of the Binary.
func (b Binary) Bytes() []byte {
	return []byte(b)
}

// String returns the string representation of the Binary.
func (b Binary) String() string {
	return "***BINARY DATA***"
}

func (b Binary) typeMarker() byte {
	return 'b'
}

func (b Binary) marshal() []byte {
	return []byte(b)
}

func (b *Binary) unmarshal(data []byte) error {
	*b = Binary(data)
	return nil
}
