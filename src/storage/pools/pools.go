package pools

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
)

// defaultBufferPool provides reusable byte buffers to reduce allocations
var defaultBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// smallBufferPool for small operations
var smallBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

// largeBufferPool for large operations (entities with content)
var largeBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 65536)) // 64KB
	},
}

// stringSlicePool provides reusable string slices
var stringSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, 32)
		return &s
	},
}

// byteSlicePool provides reusable byte slices
var byteSlicePool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 4096)
		return &b
	},
}

// EncoderWrapper wraps a JSON encoder with a buffer for pooling
type EncoderWrapper struct {
	Encoder *json.Encoder
	Buffer  *bytes.Buffer
}

// DecoderWrapper wraps a JSON decoder for pooling
type DecoderWrapper struct {
	Decoder *json.Decoder
}

// encoderPool provides reusable JSON encoder wrappers
var encoderPool = sync.Pool{
	New: func() interface{} {
		buf := &bytes.Buffer{}
		return &EncoderWrapper{
			Encoder: json.NewEncoder(buf),
			Buffer:  buf,
		}
	},
}

// decoderPool provides reusable JSON decoder wrappers  
var decoderPool = sync.Pool{
	New: func() interface{} {
		return &DecoderWrapper{
			Decoder: json.NewDecoder(strings.NewReader("")),
		}
	},
}

// stringBuilderPool provides reusable string builders
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := defaultBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 1024*1024 { // Don't pool buffers > 1MB
		return
	}
	defaultBufferPool.Put(buf)
}

// GetLargeBuffer gets a large buffer from the pool
func GetLargeBuffer() *bytes.Buffer {
	buf := largeBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutLargeBuffer returns a large buffer to the pool
func PutLargeBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 10*1024*1024 { // Don't pool buffers > 10MB
		return
	}
	largeBufferPool.Put(buf)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() *[]string {
	s := stringSlicePool.Get().(*[]string)
	*s = (*s)[:0]
	return s
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(s *[]string) {
	if cap(*s) > 1024 { // Don't pool huge slices
		return
	}
	stringSlicePool.Put(s)
}

// GetByteSlice gets a byte slice from the pool
func GetByteSlice() *[]byte {
	b := byteSlicePool.Get().(*[]byte)
	*b = (*b)[:0]
	return b
}

// PutByteSlice returns a byte slice to the pool
func PutByteSlice(b *[]byte) {
	if cap(*b) > 1024*1024 { // Don't pool slices > 1MB
		return
	}
	byteSlicePool.Put(b)
}

// GetStringBuilder gets a string builder from the pool
func GetStringBuilder() *strings.Builder {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	stringBuilderPool.Put(sb)
}

// GetJSONDecoder gets a JSON decoder wrapper from the pool
func GetJSONDecoder() *DecoderWrapper {
	return decoderPool.Get().(*DecoderWrapper)
}

// PutJSONDecoder returns a JSON decoder wrapper to the pool
func PutJSONDecoder(dec *DecoderWrapper) {
	decoderPool.Put(dec)
}

// GetJSONEncoder gets a JSON encoder wrapper from the pool
func GetJSONEncoder() *EncoderWrapper {
	wrapper := encoderPool.Get().(*EncoderWrapper)
	// Reset the buffer for reuse
	wrapper.Buffer.Reset()
	return wrapper
}

// PutJSONEncoder returns a JSON encoder wrapper to the pool
func PutJSONEncoder(enc *EncoderWrapper) {
	encoderPool.Put(enc)
}