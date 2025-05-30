package pools

import (
	"bytes"
	"sync"
	"testing"
)

func BenchmarkBufferPooling(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			buf.WriteString("test data for benchmarking buffer pools")
			for j := 0; j < 100; j++ {
				buf.WriteString("additional data")
			}
			PutBuffer(buf)
		}
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			buf.WriteString("test data for benchmarking buffer pools")
			for j := 0; j < 100; j++ {
				buf.WriteString("additional data")
			}
		}
	})
}

func BenchmarkStringSlicePooling(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s := GetStringSlice()
			for j := 0; j < 20; j++ {
				*s = append(*s, "tag:value")
			}
			PutStringSlice(s)
		}
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s := make([]string, 0, 32)
			for j := 0; j < 20; j++ {
				s = append(s, "tag:value")
			}
		}
	})
}

func TestBufferPoolConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	concurrency := 100
	iterations := 1000
	
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				buf := GetBuffer()
				buf.WriteString("concurrent test")
				PutBuffer(buf)
			}
		}()
	}
	
	wg.Wait()
}

func TestBufferPoolSizeLimits(t *testing.T) {
	// Test that large buffers are not pooled
	largeBuf := bytes.NewBuffer(make([]byte, 0, 2*1024*1024)) // 2MB
	PutBuffer(largeBuf)
	
	// Get a new buffer and verify it's not the large one
	newBuf := GetBuffer()
	if newBuf.Cap() > 1024*1024 {
		t.Errorf("Pool returned a buffer larger than expected: %d bytes", newBuf.Cap())
	}
	PutBuffer(newBuf)
}

func TestByteSlicePool(t *testing.T) {
	// Get a slice
	b := GetByteSlice()
	if b == nil {
		t.Fatal("GetByteSlice returned nil")
	}
	if len(*b) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(*b))
	}
	
	// Use the slice
	*b = append(*b, []byte("test data")...)
	
	// Put it back
	PutByteSlice(b)
	
	// Get another and verify it's been reset
	b2 := GetByteSlice()
	if len(*b2) != 0 {
		t.Errorf("Pool returned non-empty slice: %d bytes", len(*b2))
	}
	PutByteSlice(b2)
}