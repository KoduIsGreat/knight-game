package common

import "io"

type NetworkMessageEncoder interface {
	Encode() []byte
	EncodeTo(io.Writer) error
}

type NetworkMessageDecoder interface {
	Decode([]byte, any) error
	DecodeFrom(io.Reader, any) error
}
