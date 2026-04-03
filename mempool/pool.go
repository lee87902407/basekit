package mempool

func (p *BucketedPool) Get(size int) []byte {
	if size <= 0 {
		return nil
	}
	bucket := p.Bucket(size)
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
		if p.zeroOnGet {
			clear(buf)
		}
		p.onGet(size, bucket, true)
		return buf[:size]
	}
	buf := make([]byte, bucket)
	p.onGet(size, bucket, false)
	return buf[:size]
}

func (p *BucketedPool) Put(buf []byte) {
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
	if p.zeroOnPut {
		clear(full)
	}
	class.Pool.Put(full[:0])
	p.onPut(capacity, capacity, true)
}
