package secret_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/secret"
)

func TestText(t *testing.T) {
	t.Parallel()

	note := "my secret note"
	s := secret.NewText(note)

	data := s.Marshal()
	assert.NotEmpty(t, data)

	unmarshaled, err := secret.Unmarshal(data)
	require.NoError(t, err)

	v := unmarshaled.Value()
	txt := v.(*secret.Text)

	assert.Equal(t, note, txt.String())
}

func TestBinary(t *testing.T) {
	t.Parallel()

	bin := make([]byte, 1024)

	_, err := rand.Read(bin)
	require.NoError(t, err)

	s := secret.NewBinary(bin)

	data := s.Marshal()
	assert.NotEmpty(t, data)

	unmarshaled, err := secret.Unmarshal(data)
	require.NoError(t, err)

	v := unmarshaled.Value()
	got := v.(*secret.Binary)

	assert.Equal(t, bin, got.Bytes())
}

func TestPassword(t *testing.T) {
	t.Parallel()

	user := "username@localhost"
	pwd := "my secret password"
	s := secret.NewPassword(user, pwd)

	data := s.Marshal()
	assert.NotEmpty(t, data)

	unmarshaled, err := secret.Unmarshal(data)
	require.NoError(t, err)

	v := unmarshaled.Value()
	p := v.(*secret.Password)

	assert.Equal(t, user, p.Username)
	assert.Equal(t, pwd, p.Password)
}

func TestCard(t *testing.T) {
	t.Parallel()

	card := secret.NewCard("1234 5678 9012 3456", "12/22", "123")

	data := card.Marshal()
	assert.NotEmpty(t, data)

	unmarshaled, err := secret.Unmarshal(data)
	require.NoError(t, err)

	v := unmarshaled.Value()
	c := v.(*secret.Card)

	assert.Equal(t, "1234 5678 9012 3456", c.Number)
	assert.Equal(t, "12/22", c.Expiry)
	assert.Equal(t, "123", c.CVV)
}

func TestMetadate(t *testing.T) {
	t.Parallel()

	s := secret.NewText("")
	s.SetMetadataValue("key", "value")

	data := s.Marshal()
	assert.NotEmpty(t, data)

	unmarshaled, err := secret.Unmarshal(data)
	require.NoError(t, err)

	k, ok := unmarshaled.GetMetadataValue("key")
	assert.True(t, ok)
	assert.Equal(t, "value", k)

	k, ok = unmarshaled.GetMetadataValue("key2")
	assert.False(t, ok)
	assert.Empty(t, k)

	s2 := secret.NewBinary(nil)
	s2.SetMetadata(unmarshaled.Metadata())

	k, ok = s2.GetMetadataValue("key")
	assert.True(t, ok)
	assert.Equal(t, "value", k)
}
