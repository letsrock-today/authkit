package seekingbuffer

import "io"

// SeekingBuffer is an io.ReadSeeker wrapper around the byte slice.
type seekingBuffer struct {
	b    []byte
	off  int64
	fill func() ([]byte, error)
}

// New creates an empty buffer. Function fill used to lazily populate buffer at
// first Read or Seek invocation.
func New(fill func() ([]byte, error)) io.ReadSeeker {
	return &seekingBuffer{
		b:    nil,
		off:  0,
		fill: fill,
	}
}

func (b *seekingBuffer) Read(p []byte) (n int, err error) {
	if b.b == nil {
		b.b, err = b.fill()
		if err != nil {
			return 0, err
		}

	}
	if b.off >= int64(len(b.b)) {
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, b.b[b.off:])
	b.off += int64(n)
	return
}

func (b *seekingBuffer) Seek(offset int64, whence int) (n int64, err error) {
	if b.b == nil {
		b.b, err = b.fill()
		if err != nil {
			return 0, err
		}

	}
	switch whence {
	case 0:
		b.off = offset
	case 1:
		b.off = b.off + offset
	case 2:
		b.off = int64(len(b.b)) - offset
	}
	return b.off, nil
}
