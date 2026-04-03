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
