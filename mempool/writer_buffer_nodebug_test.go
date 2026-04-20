//go:build !debug

package mempool

import "testing"

// TestWriterBufferUseAfterToReaderStillPanicsWithoutDebug 验证在非 debug 模式下，
// WriterBuffer 转换为 ReaderBuffer 后再使用仍然会 panic。
// 与旧 Buffer 行为不同，新模型在非 debug 模式下也禁止使用已释放对象。
func TestWriterBufferUseAfterToReaderStillPanicsWithoutDebug(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	w.ToReaderBuffer()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on writer use after transfer without debug")
		}
	}()

	w.AppendByte('x')
}

// TestScopeCloseDoesNotDoubleReturnTransferredBuffer 验证 Scope.Close 不会重复归还
// 已转移的缓冲区。WriterBuffer 转换为 ReaderBuffer 后，缓冲区由 ReaderBuffer 管理，
// Scope.Close 应只归还一次。
func TestScopeCloseDoesNotDoubleReturnTransferredBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	r := w.ToReaderBuffer()
	wantCap := r.Cap()
	scope.Close()

	// 验证缓冲区已归还到池中
	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}

	// 验证 ReaderBuffer 在 Scope 关闭后不可用
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader use after scope close without debug")
		}
	}()

	_ = r.Bytes()
}
