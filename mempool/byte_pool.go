package mempool

import "sync"

type PoolItem struct {
	Cap  int
	Pool sync.Pool
}

type BytePool struct {
	items        []PoolItem
	maxPooledCap int
	stats        StatsCollector
}

func (p *BytePool) NewScope() *Scope {
	return &Scope{pool: p}
}

func New(opts Options) *BytePool {
	items := make([]PoolItem, len(opts.Buckets))
	for i, bucketsSize := range opts.Buckets {
		items[i] = PoolItem{Cap: bucketsSize}
	}
	return &BytePool{
		items:        items,
		maxPooledCap: opts.MaxPooledCap,
		stats:        opts.Stats,
	}
}

func (p *BytePool) bucket(size int) int {
	if size <= 0 {
		return 0
	}
	for i := range p.items {
		if size <= p.items[i].Cap {
			return p.items[i].Cap
		}
	}
	return size
}

func (p *BytePool) MaxPooledCap() int {
	return p.maxPooledCap
}

func (p *BytePool) classByCap(capacity int) *PoolItem {

	for i := range p.items {
		if p.items[i].Cap == capacity {
			return &p.items[i]
		}
	}
	return nil
}

func (p *BytePool) onGet(size int, bucket int, pooled bool) {
	if p.stats != nil {
		p.stats.OnGet(size, bucket, pooled)
	}
}

func (p *BytePool) onPut(capacity int, bucket int, pooled bool) {
	if p.stats != nil {
		p.stats.OnPut(capacity, bucket, pooled)
	}
}

func (p *BytePool) onDrop(capacity int, reason string) {
	if p.stats != nil {
		p.stats.OnDrop(capacity, reason)
	}
}

func (p *BytePool) get(size int) []byte {
	if size <= 0 {
		return nil
	}
	bucket := p.bucket(size)
	if bucket == 0 {
		return nil
	}
	if bucket > p.maxPooledCap {
		buf := make([]byte, size)
		p.onGet(size, bucket, false)
		return buf
	}
	class := p.classByCap(bucket)
	if class == nil {
		buf := make([]byte, size)
		p.onGet(size, bucket, false)
		return buf
	}
	if v := class.Pool.Get(); v != nil {
		buf := v.([]byte)
		p.onGet(size, bucket, true)
		return buf[:size]
	}
	buf := make([]byte, bucket)
	p.onGet(size, bucket, false)
	return buf[:size]
}

func (p *BytePool) put(buf []byte) {
	if buf == nil {
		return
	}
	capacity := cap(buf)
	if capacity == 0 || capacity > p.maxPooledCap {
		p.onDrop(capacity, "oversize_or_zero")
		return
	}
	class := p.classByCap(capacity)
	if class == nil {
		p.onDrop(capacity, "non_bucket_capacity")
		return
	}
	full := buf[:capacity]
	class.Pool.Put(full[:0])
	p.onPut(capacity, capacity, true)
}
