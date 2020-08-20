package kaifa

import "io"

type Reader interface {
	ReadFrame() ([]byte, error)
}

type reader struct {
	r   io.Reader
	buf []byte
}

func NewReader(r io.Reader) Reader {
	return &reader{
		r: r,
	}
}

func (r *reader) ReadFrame() ([]byte, error) {
	buf := make([]byte, 4096)
	for {
		n, err := r.r.Read(buf)
		if err != nil {
			return nil, err
		}
		r.buf = append(r.buf, buf[0:n]...)
		if fr, err := r.tryFrame(); len(fr) > 0 || err != nil {
			return fr, err
		}
	}
}

func (r *reader) tryFrame() ([]byte, error) {
	// Forward to next frame
	// useful in case we start reading in the middle of a frame
	for len(r.buf) > 1 && r.buf[0] != frameTag && (r.buf[1]&frameFormatMask) != frameFormat {
		r.buf = r.buf[1:]
	}
	if len(r.buf) < 3 {
		return nil, nil
	}
	length := (uint16(r.buf[1]&frameLengthMask) << 8) + uint16(r.buf[2])

	if (len(r.buf) - 2) < int(length) {
		// Read more data before we have the full frame
		return nil, nil
	}
	fr := r.buf[1 : int(length)+1]
	r.buf = r.buf[int(length)+2:]
	return fr, nil
}
