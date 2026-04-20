//go:build debug

package mempool

import "testing"

// TestWriterBufferUseAfterToReaderPanics 验证在 debug 模式下，
// WriterBuffer 转换为 ReaderBuffer 后再使用会 panic。
func TestWriterBufferUseAfterToReaderPanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	w.ToReaderBuffer()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on writer use after transfer")
		}
	}()

	w.AppendByte('x')
}

// TestReaderBufferUseAfterScopeClosePanics 验证在 debug 模式下，
// Scope 关闭后再使用 ReaderBuffer 会 panic。
func TestReaderBufferUseAfterScopeClosePanics(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	r := w.ToReaderBuffer()
	scope.Close()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader use after scope close")
		}
	}()

	_ = r.Len()
}
