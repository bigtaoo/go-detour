package aligned

import (
	"encoding/binary"
	"io"
	"reflect"
)

// A Reader manages its embedded io.Reader so as to guarantee that every read
// operation consumes an amount of bytes that is a multiple of the provided
// alignment.
type Reader struct {
	r      io.Reader        // Reader to which calls are forwarded
	align  int              // byte alignment
	padbuf []byte           // buffer used to discard padding
	order  binary.ByteOrder // byte order to use when unpacking
}

// NewReader creates a Reader that performs aligned read operation on r.
func NewReader(r io.Reader, align int, order binary.ByteOrder) *Reader {
	return &Reader{
		r:      r,
		align:  align,
		padbuf: make([]byte, align),
		order:  order,
	}
}

// Read reads len(b) bytes of data into b, afterwards it consumes eventual
// padding bytes from the underlying data stream, in order to respect the
// specified aligment.
//
// It returns the total number of bytes read and any error encountered. If
// err == nil, it is guaranteed that the number of bytes consumed from the
// underlying stream is a multiple of alignment. io.ErrUnexpectedEOF is
// returned if the underlying stream doesn't contain enough data to fill b
// entirely, nor consumes the padding. Other read errors are forwarded.
func (ar *Reader) Read(b []byte) (n int, err error) {
	// we return 0 as the number of bytes also if it we
	// actually read n + npad bytes otherwise the error
	// in not taken into consideration, at least that's
	// what happen if the ar.r is Reader from
	// encoding/binary package.
	//
	// read the significative part of the underlying stream
	n, err = ar.r.Read(b)
	switch {
	case err != nil:
		return 0, err
	case n < len(b):
		// not enough data to fill b entirely
		return 0, io.ErrUnexpectedEOF
	}

	// compute the number of padding bytes we need
	var pad, npad int
	pad = AlignN(n, ar.align) - n
	if pad != 0 {
		// consumes (and discard) the padding bytes
		npad, err = ar.r.Read(ar.padbuf[:pad])
		switch {
		case err != nil:
			return 0, err
		case npad < pad:
			// not enough padding to keep the data stream aligned
			return 0, io.ErrUnexpectedEOF
		}
	}
	return n + pad, nil
}

// ReadSlice reads len(s) aligned elements from the underlying data stream,
// into s.
//
// s must be an allocated slice. It returns an error if the slice can't be
// filled entirely. The number of bytes of data consumed for each element of
// the slice is equal to the size of one element plus the padding required to
// have each element alignment on the specified wordsize.
func (ar *Reader) ReadSlice(s interface{}) error {
	var elem reflect.Value
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Slice {
		elem = rv
	} else if rv.Kind() == reflect.Ptr {
		elem = rv.Elem()
	}
	// elem := reflect.ValueOf(s).Elem()
	for idx := 0; idx < elem.Len(); idx++ {
		err := binary.Read(ar, ar.order, elem.Index(idx).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadVal reads aligned binary data from the embedded io.Reader, into v.
//
// see binary.Read from encoding/binary package for more information.
func (ar *Reader) ReadVal(v interface{}) error {
	return binary.Read(ar, ar.order, v)
}
