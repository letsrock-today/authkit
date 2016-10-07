package seekingbuffer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekingBuffer(t *testing.T) {
	assert := assert.New(t)

	executed := false
	fill := func() ([]byte, error) {
		executed = true
		return []byte("abcdefghijklmnopqrstuvwxyz"), nil
	}
	b := New(fill)
	assert.NotZero(b)
	assert.False(executed)

	b1 := make([]byte, 5)

	i, err := b.Read(b1)
	assert.NoError(err)
	assert.Equal(5, i)
	assert.True(executed)
	assert.Equal("abcde", string(b1))

	j, err := b.Seek(10, 0)
	assert.NoError(err)
	assert.Equal(int64(10), j)

	i, err = b.Read(b1)
	assert.NoError(err)
	assert.Equal(5, i)
	assert.Equal("klmno", string(b1))

	j, err = b.Seek(5, 1)
	assert.NoError(err)
	assert.Equal(int64(20), j)

	i, err = b.Read(b1)
	assert.NoError(err)
	assert.Equal(5, i)
	assert.Equal("uvwxy", string(b1))

	j, err = b.Seek(3, 2)
	assert.NoError(err)
	assert.Equal(int64(23), j)

	i, err = b.Read(b1)
	assert.NoError(err)
	assert.Equal(3, i)
	assert.Equal("xyz", string(b1[:i]))
}

func TestSeekFirst(t *testing.T) {
	assert := assert.New(t)

	executed := false
	fill := func() ([]byte, error) {
		executed = true
		return []byte("abcdefghijklmnopqrstuvwxyz"), nil
	}
	b := New(fill)
	assert.NotZero(b)
	assert.False(executed)

	j, err := b.Seek(10, 1)
	assert.NoError(err)
	assert.Equal(int64(10), j)
	assert.True(executed)

	b1 := make([]byte, 5)

	i, err := b.Read(b1)
	assert.NoError(err)
	assert.Equal(5, i)
	assert.Equal("klmno", string(b1))
}

func TestFillErrors(t *testing.T) {
	assert := assert.New(t)

	fill := func() ([]byte, error) {
		return nil, errors.New("error")
	}
	b := New(fill)
	_, err := b.Seek(10, 1)
	assert.Error(err)

	b1 := make([]byte, 5)

	_, err = b.Read(b1)
	assert.Error(err)
}

func TestEOF(t *testing.T) {
	assert := assert.New(t)

	fill := func() ([]byte, error) {
		return []byte("abc"), nil
	}
	b := New(fill)
	_, err := b.Seek(10, 1)
	assert.NoError(err)

	b1 := make([]byte, 5)

	_, err = b.Read(b1)
	assert.Error(err)
}
