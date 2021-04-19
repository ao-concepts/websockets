package websockets

import "sync"

type Batch struct {
	lock sync.Mutex
	data []Payload
}

// NewBatch constructor
func NewBatch() *Batch {
	return &Batch{
		lock: sync.Mutex{},
	}
}

// AddPayload add payload to the batch
func (b *Batch) AddPayload(payload Payload) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.data = append(b.data, payload)
}

// GetDataAndRemoveAll get the current payload batch and remove all data from the batch
func (b *Batch) GetDataAndRemoveAll() []Payload {
	b.lock.Lock()
	defer b.lock.Unlock()

	data := b.data
	b.data = nil

	return data
}
