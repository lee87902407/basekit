package mempool

import "sync"

type SizeClass struct {
	Cap  int
	Pool sync.Pool
}

type BucketedPool struct {
	classes      []SizeClass
	maxPooledCap int
	zeroOnPut    bool
	zeroOnGet    bool
	stats        StatsCollector
}

func New(opts Options) *BucketedPool {
	classes := make([]SizeClass, len(opts.Buckets))
	for i, cap := range opts.Buckets {
		classes[i] = SizeClass{Cap: cap}
	}
	return &BucketedPool{
		classes:      classes,
		maxPooledCap: opts.MaxPooledCap,
		zeroOnPut:    opts.ZeroOnPut,
		zeroOnGet:    opts.ZeroOnGet,
		stats:        opts.Stats,
	}
}

func (p *BucketedPool) Bucket(size int) int {
	if size <= 0 {
		return 0
	}
	for i := range p.classes {
		if size <= p.classes[i].Cap {
			return p.classes[i].Cap
		}
	}
	return size
}

func (p *BucketedPool) MaxPooledCap() int {
	return p.maxPooledCap
}

func (p *BucketedPool) classByCap(capacity int) *SizeClass {
	for i := range p.classes {
		if p.classes[i].Cap == capacity {
			return &p.classes[i]
		}
	}
	return nil
}

func (p *BucketedPool) onGet(size int, bucket int, pooled bool) {
	if p.stats != nil {
		p.stats.OnGet(size, bucket, pooled)
	}
}

func (p *BucketedPool) onPut(capacity int, bucket int, pooled bool) {
	if p.stats != nil {
		p.stats.OnPut(capacity, bucket, pooled)
	}
}

func (p *BucketedPool) onDrop(capacity int, reason string) {
	if p.stats != nil {
		p.stats.OnDrop(capacity, reason)
	}
}
