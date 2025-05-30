package pools

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
)

// BufferPool provides reusable byte buffers to reduce allocations
var BufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// SmallBufferPool for small operations
var SmallBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

// LargeBufferPool for large operations (entities with content)
var LargeBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 65536)) // 64KB
	},
}

// StringSlicePool provides reusable string slices
var StringSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, 32)
		return &s
	},
}

// ByteSlicePool provides reusable byte slices
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 4096)
		return &b
	},
}

// DecoderPool provides reusable JSON decoders
var DecoderPool = sync.Pool{
	New: func() interface{} {
		return json.NewDecoder(nil)
	},
}

// EncoderPool provides reusable JSON encoders
var EncoderPool = sync.Pool{
	New: func() interface{} {
		return json.NewEncoder(nil)
	},
}

// StringBuilderPool provides reusable string builders
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 1024*1024 { // Don't pool buffers > 1MB
		return
	}
	BufferPool.Put(buf)
}

// GetLargeBuffer gets a large buffer from the pool
func GetLargeBuffer() *bytes.Buffer {
	buf := LargeBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutLargeBuffer returns a large buffer to the pool
func PutLargeBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 10*1024*1024 { // Don't pool buffers > 10MB
		return
	}
	LargeBufferPool.Put(buf)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() *[]string {
	s := StringSlicePool.Get().(*[]string)
	*s = (*s)[:0]
	return s
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(s *[]string) {
	if cap(*s) > 1024 { // Don't pool huge slices
		return
	}
	StringSlicePool.Put(s)
}

// GetByteSlice gets a byte slice from the pool
func GetByteSlice() *[]byte {
	b := ByteSlicePool.Get().(*[]byte)
	*b = (*b)[:0]
	return b
}

// PutByteSlice returns a byte slice to the pool
func PutByteSlice(b *[]byte) {
	if cap(*b) > 1024*1024 { // Don't pool slices > 1MB
		return
	}
	ByteSlicePool.Put(b)
}

// GetStringBuilder gets a string builder from the pool
func GetStringBuilder() *strings.Builder {
	sb := StringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	StringBuilderPool.Put(sb)
}

// GetJSONDecoder gets a JSON decoder from the pool
func GetJSONDecoder() *json.Decoder {
	return DecoderPool.Get().(*json.Decoder)
}

// PutJSONDecoder returns a JSON decoder to the pool
func PutJSONDecoder(dec *json.Decoder) {
	DecoderPool.Put(dec)
}

// GetJSONEncoder gets a JSON encoder from the pool
func GetJSONEncoder() *json.Encoder {
	return EncoderPool.Get().(*json.Encoder)
}

// PutJSONEncoder returns a JSON encoder to the pool
func PutJSONEncoder(enc *json.Encoder) {
	EncoderPool.Put(enc)
}