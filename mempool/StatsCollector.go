package mempool

type StatsCollector interface {
	OnGet(size int, bucket int, pooled bool)
	OnPut(capacity int, bucket int, pooled bool)
	OnDrop(capacity int, reason string)
}
