# Mempool Buffer Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `mempool.Buffer` 拆分为 `WriterBuffer` 与 `ReaderBuffer`，把生命周期管理收敛到 `Scope`，并同步更新测试、示例与文档。

**Architecture:** `mempool` 改为显式双类型模型：`Scope.NewWriterBuffer()` 创建可写对象，`WriterBuffer.ToReaderBuffer()` 转移底层 `[]byte` 所有权到只读对象，`Scope.Close()` 统一负责归还 writer / reader 与裸 `[]byte`。原公开 `Buffer`、`NewBuffer`、`Release()` 全部移除；debug 与非 debug 下都不再允许失效对象自动恢复。

**Tech Stack:** Go 1.25、标准库 testing、build tags (`debug` / `!debug`)

---

### Task 1: 建立 WriterBuffer 与 ReaderBuffer 的最小类型骨架

**Files:**
- Create: `mempool/writer_buffer.go`
- Create: `mempool/reader_buffer.go`
- Modify: `mempool/scope.go`
- Test: `mempool/writer_buffer_test.go`

- [ ] **Step 1: Write the failing test**

```go
package mempool

import "testing"

func TestScopeNewWriterBufferCreatesWriter(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(64)

	if w == nil {
		t.Fatal("writer buffer should not be nil")
	}
	if w.Released() {
		t.Fatal("new writer buffer should not be released")
	}
	if w.Len() != 64 {
		t.Fatalf("writer len = %d, want 64", w.Len())
	}
	if w.Cap() < 64 {
		t.Fatalf("writer cap = %d, want >= 64", w.Cap())
	}
	if len(scope.writers) != 1 {
		t.Fatalf("scope writers = %d, want 1", len(scope.writers))
	}
}

func TestWriterBufferToReaderBufferTransfersOwnership(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(8)
	w.Reset()
	w.Append([]byte("hello"))

	r := w.ToReaderBuffer()

	if r == nil {
		t.Fatal("reader buffer should not be nil")
	}
	if !w.Released() {
		t.Fatal("writer should be released after transfer")
	}
	if r.Released() {
		t.Fatal("reader should be active after transfer")
	}
	if string(r.Bytes()) != "hello" {
		t.Fatalf("reader bytes = %q, want %q", string(r.Bytes()), "hello")
	}
	if len(scope.readers) != 1 {
		t.Fatalf("scope readers = %d, want 1", len(scope.readers))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./mempool -run 'TestScopeNewWriterBufferCreatesWriter|TestWriterBufferToReaderBufferTransfersOwnership'`
Expected: FAIL with undefined `NewWriterBuffer` / `WriterBuffer` / `ReaderBuffer`

- [ ] **Step 3: Write minimal implementation**

```go
package mempool

type WriterBuffer struct {
	buf      []byte
	pool     BytePool
	scope    *Scope
	released bool
}

func (b *WriterBuffer) Bytes() []byte  { b.mustUsable(); return b.buf }
func (b *WriterBuffer) Len() int       { b.mustUsable(); return len(b.buf) }
func (b *WriterBuffer) Cap() int       { b.mustUsable(); return cap(b.buf) }
func (b *WriterBuffer) Released() bool { return b.released }

func (b *WriterBuffer) ToReaderBuffer() *ReaderBuffer {
	b.mustUsable()
	r := &ReaderBuffer{buf: b.buf, pool: b.pool, released: false}
	b.buf = nil
	b.released = true
	if b.scope != nil {
		b.scope.readers = append(b.scope.readers, r)
	}
	return r
}

func (b *WriterBuffer) releaseToPool() {
	if b.released {
		return
	}
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
```

```go
package mempool

type ReaderBuffer struct {
	buf      []byte
	pool     BytePool
	released bool
}

func (b *ReaderBuffer) Bytes() []byte  { b.mustUsable(); return b.buf }
func (b *ReaderBuffer) Len() int       { b.mustUsable(); return len(b.buf) }
func (b *ReaderBuffer) Cap() int       { b.mustUsable(); return cap(b.buf) }
func (b *ReaderBuffer) Released() bool { return b.released }

func (b *ReaderBuffer) releaseToPool() {
	if b.released {
		return
	}
	b.pool.Put(b.buf)
	b.buf = nil
	b.released = true
}
```

```go
package mempool

type Scope struct {
	pool    BytePool
	writers []*WriterBuffer
	readers []*ReaderBuffer
	raws    [][]byte
	closed  bool
}

func NewScope(pool BytePool) *Scope {
	return &Scope{pool: pool}
}

func (s *Scope) Get(size int) []byte {
	buf := s.pool.Get(size)
	s.raws = append(s.raws, buf)
	return buf
}

func (s *Scope) NewWriterBuffer(size int) *WriterBuffer {
	b := &WriterBuffer{buf: s.pool.Get(size), pool: s.pool, scope: s}
	s.writers = append(s.writers, b)
	return b
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./mempool -run 'TestScopeNewWriterBufferCreatesWriter|TestWriterBufferToReaderBufferTransfersOwnership'`
Expected: PASS

### Task 2: 迁移 WriterBuffer 写接口与 ReaderBuffer 只读接口

**Files:**
- Modify: `mempool/writer_buffer.go`
- Modify: `mempool/reader_buffer.go`
- Modify: `mempool/writer_buffer_test.go`

- [ ] **Step 1: Write the failing test**

```go
package mempool

import (
	"bytes"
	"testing"
)

func TestWriterBufferAppendGrowsThroughPoolWhenCapacityInsufficient(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(510)
	w.Reset()
	w.Append(bytes.Repeat([]byte{'a'}, 510))

	oldCap := w.Cap()
	w.Append(bytes.Repeat([]byte{'b'}, 10))

	if w.Cap() <= oldCap {
		t.Fatalf("writer cap did not grow, old=%d new=%d", oldCap, w.Cap())
	}
	if w.Len() != 520 {
		t.Fatalf("writer len = %d, want 520", w.Len())
	}
	if w.Cap() != 1024 {
		t.Fatalf("writer cap = %d, want 1024", w.Cap())
	}
}

func TestWriterBufferCloneCreatesDetachedCopy(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(10)
	w.Reset()
	w.Append([]byte("hello"))

	dup := w.Clone()
	w.Append([]byte("-world"))

	if string(dup) != "hello" {
		t.Fatalf("clone = %q, want hello", string(dup))
	}
}

func TestReaderBufferCloneCreatesDetachedCopy(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(10)
	w.Reset()
	w.Append([]byte("hello"))
	r := w.ToReaderBuffer()

	dup := r.Clone()
	if string(dup) != "hello" {
		t.Fatalf("clone = %q, want hello", string(dup))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./mempool -run 'TestWriterBufferAppendGrowsThroughPoolWhenCapacityInsufficient|TestWriterBufferCloneCreatesDetachedCopy|TestReaderBufferCloneCreatesDetachedCopy'`
Expected: FAIL with missing `Reset` / `Append` / `Clone` / `EnsureCapacity`

- [ ] **Step 3: Write minimal implementation**

```go
package mempool

func (b *WriterBuffer) Reset() {
	b.mustUsable()
	b.buf = b.buf[:0]
}

func (b *WriterBuffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

func (b *WriterBuffer) DetachedCopy() []byte {
	return b.Clone()
}

func (b *WriterBuffer) EnsureCapacity(additional int) {
	b.mustUsable()
	if additional <= 0 {
		return
	}
	need := len(b.buf) + additional
	if need <= cap(b.buf) {
		return
	}
	next := b.pool.Get(need)
	copy(next, b.buf)
	b.pool.Put(b.buf)
	b.buf = next[:len(b.buf)]
}

func (b *WriterBuffer) Resize(n int) {
	b.mustUsable()
	if n <= cap(b.buf) {
		b.buf = b.buf[:n]
		return
	}
	next := b.pool.Get(n)
	copy(next, b.buf)
	b.pool.Put(b.buf)
	b.buf = next[:n]
}

func (b *WriterBuffer) Append(p []byte) {
	b.mustUsable()
	b.EnsureCapacity(len(p))
	start := len(b.buf)
	b.buf = b.buf[:start+len(p)]
	copy(b.buf[start:], p)
}

func (b *WriterBuffer) AppendByte(v byte) {
	b.mustUsable()
	b.EnsureCapacity(1)
	b.buf = append(b.buf, v)
}
```

```go
package mempool

func (b *ReaderBuffer) Clone() []byte {
	b.mustUsable()
	dup := make([]byte, len(b.buf))
	copy(dup, b.buf)
	return dup
}

func (b *ReaderBuffer) DetachedCopy() []byte {
	return b.Clone()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./mempool -run 'TestWriterBufferAppendGrowsThroughPoolWhenCapacityInsufficient|TestWriterBufferCloneCreatesDetachedCopy|TestReaderBufferCloneCreatesDetachedCopy'`
Expected: PASS

### Task 3: 接入 debug / non-debug 失效检查并完成 Scope.Close

**Files:**
- Create: `mempool/writer_buffer_debug_checks.go`
- Create: `mempool/writer_buffer_nodebug_checks.go`
- Create: `mempool/reader_buffer_debug_checks.go`
- Create: `mempool/reader_buffer_nodebug_checks.go`
- Modify: `mempool/scope.go`
- Create: `mempool/writer_buffer_debug_test.go`
- Create: `mempool/writer_buffer_nodebug_test.go`

- [ ] **Step 1: Write the failing tests**

```go
//go:build debug

package mempool

import "testing"

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
```

```go
//go:build !debug

package mempool

import "testing"

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

func TestScopeCloseDoesNotDoubleReturnTransferredBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	r := w.ToReaderBuffer()
	wantCap := r.Cap()
	scope.Close()

	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on reader use after scope close without debug")
		}
	}()

	_ = r.Bytes()
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -tags debug ./mempool -run 'TestWriterBufferUseAfterToReaderPanics|TestReaderBufferUseAfterScopeClosePanics' && go test ./mempool -run 'TestWriterBufferUseAfterToReaderStillPanicsWithoutDebug|TestScopeCloseDoesNotDoubleReturnTransferredBuffer'`
Expected: FAIL because `mustUsable` / `Scope.Close` are incomplete or still use old behavior

- [ ] **Step 3: Write minimal implementation**

```go
//go:build debug

package mempool

func (b *WriterBuffer) mustUsable() {
	if b.released {
		panic("mempool: writer buffer is released")
	}
}

func (b *ReaderBuffer) mustUsable() {
	if b.released {
		panic("mempool: reader buffer is released")
	}
}
```

```go
//go:build !debug

package mempool

func (b *WriterBuffer) mustUsable() {
	if b.released {
		panic("mempool: writer buffer is released")
	}
}

func (b *ReaderBuffer) mustUsable() {
	if b.released {
		panic("mempool: reader buffer is released")
	}
}
```

```go
package mempool

func (s *Scope) Track(buf []byte) {
	s.raws = append(s.raws, buf)
}

func (s *Scope) Close() {
	if s.closed {
		return
	}
	for i := range s.writers {
		s.writers[i].releaseToPool()
	}
	for i := range s.readers {
		s.readers[i].releaseToPool()
	}
	for i := range s.raws {
		s.pool.Put(s.raws[i])
	}
	s.closed = true
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -tags debug ./mempool -run 'TestWriterBufferUseAfterToReaderPanics|TestReaderBufferUseAfterScopeClosePanics' && go test ./mempool -run 'TestWriterBufferUseAfterToReaderStillPanicsWithoutDebug|TestScopeCloseDoesNotDoubleReturnTransferredBuffer'`
Expected: PASS

### Task 4: 移除旧 Buffer 入口并迁移现有测试

**Files:**
- Delete: `mempool/buffer.go`
- Delete: `mempool/buffer_debug_checks.go`
- Delete: `mempool/buffer_nodebug_checks.go`
- Delete: `mempool/buffer_test.go`
- Delete: `mempool/buffer_nodebug_test.go`
- Modify: `mempool/scope_test.go`
- Modify: `mempool/interface.go`

- [ ] **Step 1: Write the failing migration test**

```go
package mempool

import "testing"

func TestScopeCloseReleasesWriterBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	wantCap := w.Cap()

	scope.Close()

	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}
	if !w.Released() {
		t.Fatal("writer should be released after scope close")
	}
}
```

- [ ] **Step 2: Run targeted test to verify migration gaps**

Run: `go test ./mempool -run 'TestScopeCloseReleasesWriterBuffer'`
Expected: If old files still conflict, build or symbol errors surface and guide cleanup

- [ ] **Step 3: Remove old Buffer API and update remaining tests**

```go
package mempool

type BytePool interface {
	Get(size int) []byte
	Put(buf []byte)
	Bucket(size int) int
	MaxPooledCap() int
}

type StatsCollector interface {
	OnGet(size int, bucket int, pooled bool)
	OnPut(capacity int, bucket int, pooled bool)
	OnDrop(capacity int, reason string)
}
```

```go
package mempool

import "testing"

func TestScopeCloseStillReleasesWriterBuffer(t *testing.T) {
	pool := New(DefaultOptions())
	scope := NewScope(pool)
	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("abc"))
	wantCap := w.Cap()

	scope.Close()

	reused := pool.Get(1)
	if cap(reused) != wantCap {
		t.Fatalf("cap(reused) = %d, want %d", cap(reused), wantCap)
	}
}
```

- [ ] **Step 4: Run mempool tests after cleanup**

Run: `go test ./mempool && go test -tags debug ./mempool`
Expected: PASS

### Task 5: 更新 README、示例与设计文档

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `README.en.md`
- Modify: `docs/examples/mempool.md`
- Modify: `examples/mempool/main.go`
- Modify: `docs/designs/memory-pool-design.md`

- [ ] **Step 1: Write the failing documentation expectation**

```text
需要在文档和示例中体现以下变化，否则视为未完成：
1. `Scope.NewWriterBuffer()` 成为创建入口。
2. `WriterBuffer` 与 `ReaderBuffer` 的职责区别。
3. `ToReaderBuffer()` 转移 ownership。
4. `WriterBuffer` 转换后立即失效。
5. `Scope.Close()` 是统一释放入口。
6. 不再对外介绍公开 `Buffer.Release()` 与 `NewBuffer()`。
```

- [ ] **Step 2: Verify expectation currently fails**

Run: `rg -n "NewBuffer|Buffer\.Release|WriterBuffer|ReaderBuffer|ToReaderBuffer|NewWriterBuffer" README.md README.zh-CN.md README.en.md docs/examples/mempool.md examples/mempool/main.go docs/designs/memory-pool-design.md`
Expected: docs still mention old `Buffer` / `Release` / missing new names

- [ ] **Step 3: Write minimal documentation and example changes**

```markdown
# docs/examples/mempool.md 新增/替换要点

## 轻量 buffer 模型

`mempool` 现在区分两个对象：

1. `WriterBuffer`：由 `Scope.NewWriterBuffer()` 创建，负责写阶段的追加、扩容、重置与构建。
2. `ReaderBuffer`：由 `WriterBuffer.ToReaderBuffer()` 产生，只保留只读能力。

`ToReaderBuffer()` 会把底层 `[]byte` 的 ownership 从 writer 转移给 reader。转换完成后，原 `WriterBuffer` 立即失效，不能再继续访问。

buffer 的底层内存统一由 `Scope.Close()` 归还，不再对外暴露公开 `Release()`。
```

```go
package main

import (
	"fmt"

	"github.com/lee87902407/basekit/mempool"
)

func main() {
	pool := mempool.New(mempool.DefaultOptions())
	scope := mempool.NewScope(pool)
	defer scope.Close()

	w := scope.NewWriterBuffer(32)
	w.Reset()
	w.Append([]byte("hello mempool"))
	r := w.ToReaderBuffer()

	fmt.Printf("len=%d cap=%d text=%q\n", r.Len(), r.Cap(), string(r.Bytes()))
}
```

```markdown
# README.md / README.zh-CN.md / README.en.md 需要补充的行为说明

- `mempool` 现在提供 `WriterBuffer` / `ReaderBuffer` 双类型模型。
- writer 由 `Scope.NewWriterBuffer()` 创建。
- `ToReaderBuffer()` 会转移 ownership，writer 随即失效。
- `Scope.Close()` 统一释放底层缓冲区。
```

- [ ] **Step 4: Verify docs and example are updated**

Run: `go test ./... && rg -n "NewWriterBuffer|WriterBuffer|ReaderBuffer|ToReaderBuffer|Scope.Close|Release\(" README.md README.zh-CN.md README.en.md docs/examples/mempool.md examples/mempool/main.go docs/designs/memory-pool-design.md`
Expected: PASS, docs mention new model and no longer present old public guidance for `NewBuffer`/`Release`

### Task 6: 全量验证与残留出口检查

**Files:**
- Modify: `mempool/writer_buffer.go`
- Modify: `mempool/reader_buffer.go`
- Modify: `mempool/scope.go`
- Modify: `mempool/writer_buffer_test.go`
- Modify: `mempool/writer_buffer_debug_test.go`
- Modify: `mempool/writer_buffer_nodebug_test.go`
- Modify: `docs/examples/mempool.md`
- Modify: `examples/mempool/main.go`

- [ ] **Step 1: Run LSP diagnostics on mempool files**

Run: `lsp_diagnostics` for `mempool/writer_buffer.go`, `mempool/reader_buffer.go`, `mempool/scope.go`, `mempool/writer_buffer_test.go`, `mempool/writer_buffer_debug_test.go`, `mempool/writer_buffer_nodebug_test.go`
Expected: no errors

- [ ] **Step 2: Run full test suites**

Run: `go test ./... && go test -tags debug ./...`
Expected: PASS

- [ ] **Step 3: Check no old public API remains**

Run: `rg -n "type Buffer struct|func NewBuffer\(|func \(b \*Buffer\) Release\(" mempool`
Expected: no matches

- [ ] **Step 4: Check documentation matches new model**

```text
检查以下点：
1. README 与示例文档都提到了 `Scope.NewWriterBuffer()`。
2. README 与示例文档都说明 `ToReaderBuffer()` 转移 ownership。
3. 没有把公开 `Release()` 继续写成推荐用法。
4. 示例代码展示 writer -> reader -> scope close 的完整路径。
```
