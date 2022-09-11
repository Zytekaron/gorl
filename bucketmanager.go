package gorl

import (
	"sync"
	"time"
)

// BucketManager is a thread-safe manager for Bucket instances.
// Buckets are automatically created when they are queried if they do
// not already exist, and can be managed directly from this type.
//
// Buckets are not automatically removed when they no longer contain
// useful information (when they have fully refilled), but you can call
// Purge
type BucketManager struct {
	Limit  int64
	Burst  int64
	Refill time.Duration

	buckets   map[string]*Bucket
	bucketMux sync.RWMutex
}

func New(limit, burst int64, refill time.Duration) *BucketManager {
	return &BucketManager{
		Limit:   limit,
		Burst:   burst,
		Refill:  refill,
		buckets: make(map[string]*Bucket),
	}
}

// Get gets a bucket from the BucketManager, creating it if necessary.
func (m *BucketManager) Get(id string) *Bucket {
	return m.getOrCreate(id)
}

// Set adds a bucket to the BucketManager.
func (m *BucketManager) Set(id string, bucket *Bucket) {
	m.set(id, bucket)
}

// Delete removes a bucket from the BucketManager.
func (m *BucketManager) Delete(id string) {
	m.delete(id)
}

// CanDraw returns whether there are enough tokens remaining in the bucket to draw n.
func (m *BucketManager) CanDraw(id string, n int64) bool {
	return m.getOrCreate(id).CanDraw(n)
}

// CanDrawAt returns whether there are enough tokens remaining in the bucket to draw n.
func (m *BucketManager) CanDrawAt(id string, t time.Time, n int64) bool {
	return m.getOrCreate(id).CanDrawAt(t, n)
}

// Draw draws n tokens from the bucket, returning whether there were enough tokens
// remaining to draw without overdraft. If not, no tokens are drawn from the bucket.
func (m *BucketManager) Draw(id string, n int64) bool {
	return m.getOrCreate(id).Draw(n)
}

// DrawAt draws n tokens from the bucket, returning whether there were enough tokens
// remaining to draw without overdraft. If not, no tokens are drawn from the bucket.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens.
func (m *BucketManager) DrawAt(id string, t time.Time, n int64) bool {
	return m.getOrCreate(id).DrawAt(t, n)
}

// DrawMax attempts to draw up to n tokens, returning the number of tokens drawn.
func (m *BucketManager) DrawMax(id string, n int64) int64 {
	return m.getOrCreate(id).DrawMax(n)
}

// DrawMaxAt attempts to draw up to n tokens, returning the number of tokens drawn.
func (m *BucketManager) DrawMaxAt(id string, t time.Time, n int64) int64 {
	return m.getOrCreate(id).DrawMaxAt(t, n)
}

// ForceDraw forcefully draws a certain number of tokens and
// returns the number of remaining uses, which may be negative.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens.
// for potentially multiple refill intervals.
func (m *BucketManager) ForceDraw(id string, n int64) int64 {
	return m.getOrCreate(id).ForceDraw(n)
}

// ForceDrawAt forcefully draws a certain number of tokens and
// returns the number of remaining uses, which may be negative.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens
// for potentially multiple refill intervals.
func (m *BucketManager) ForceDrawAt(id string, t time.Time, n int64) int64 {
	return m.getOrCreate(id).ForceDrawAt(t, n)
}

// SetTokens sets the number of available tokens and sets the last update time to the current time.
func (m *BucketManager) SetTokens(id string, tokens int64) {
	m.getOrCreate(id).SetTokens(tokens)
}

// SetTokensAt sets the number of available tokens and sets the last update time to the provided time.
func (m *BucketManager) SetTokensAt(id string, t time.Time, tokens int64) {
	m.getOrCreate(id).SetTokensAt(t, tokens)
}

// Remaining returns the remaining tokens which can be drawn.
//
// If the number of tokens in the bucket is less than zero, this returns 0.
func (m *BucketManager) Remaining(id string) int64 {
	return m.getOrCreate(id).Remaining()
}

// RemainingAt returns the remaining tokens which can be drawn at the specified time.
//
// If the number of tokens in the bucket is less than zero, this returns 0.
func (m *BucketManager) RemainingAt(id string, t time.Time) int64 {
	return m.getOrCreate(id).RemainingAt(t)
}

// Tokens returns the number of tokens in the bucket.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (m *BucketManager) Tokens(id string) int64 {
	return m.getOrCreate(id).TokensAt(time.Now())
}

// TokensAt returns the number of tokens in the bucket at the specified time.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (m *BucketManager) TokensAt(id string, t time.Time) int64 {
	return m.getOrCreate(id).TokensAt(t)
}

// InferTokensAt returns the number of tokens that will be in the bucket at the
// specified time. This assumes that there will be no modifications to the bucket
// between the current and provided time, such as Draw, Reset, or SetTokens.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (m *BucketManager) InferTokensAt(id string, t time.Time) int64 {
	return m.getOrCreate(id).InferTokensAt(t)
}

// NextRefill returns the next time this bucket will refill.
//
// This method does not modify the bucket, so it may be called with times which are out of chronology.
func (m *BucketManager) NextRefill(id string) time.Time {
	return m.getOrCreate(id).NextRefillAt(time.Now())
}

// NextRefillAt returns the next time this bucket will refill, after the specified time.
//
// This method does not modify the bucket, so it may be called with times which are out of chronology.
func (m *BucketManager) NextRefillAt(id string, t time.Time) time.Time {
	return m.getOrCreate(id).NextRefillAt(t)
}

// Reset resets this bucket. The number of tokens available is reset to
// the burst quantity, and the last update time is set to the current time.
func (m *BucketManager) Reset(id string) {
	m.getOrCreate(id).Reset()
}

// ResetAt resets this bucket. The number of tokens available is reset to
// the burst quantity, and the last update time is set to the provided time.
func (m *BucketManager) ResetAt(id string, t time.Time) {
	m.getOrCreate(id).ResetAt(t)
}

// IsReset returns whether this bucket has just been created or is reset to
// a point where it can be fully drawn from up to the burst quantity.
func (m *BucketManager) IsReset(id string) bool {
	return m.getOrCreate(id).IsReset()
}

// IsResetAt returns whether this bucket has just been created or is reset to
// a point where it can be fully drawn from up to the burst quantity.
func (m *BucketManager) IsResetAt(id string, t time.Time) bool {
	return m.getOrCreate(id).IsResetAt(t)
}

// Purge removes elements which return true for IsReset, returning
// the number of buckets which were removed during the iteration.
//
// There is currently no PurgeAt counterpart because it may cause
// issues if the buckets are modified between the time that the
// purge loop starts and the time that they would be removed.
func (m *BucketManager) Purge() int {
	removed := 0

	m.bucketMux.RLock()
	for id, bucket := range m.buckets {
		if bucket.IsReset() {
			m.bucketMux.RUnlock()
			m.bucketMux.Lock()
			delete(m.buckets, id)
			m.bucketMux.Unlock()
			m.bucketMux.RLock()
		}
	}
	m.bucketMux.RUnlock()

	return removed
}

func (m *BucketManager) getOrCreate(id string) *Bucket {
	if bucket, ok := m.get(id); ok {
		return bucket
	}

	bucket := NewBucket(m.Limit, m.Burst, m.Refill)
	m.set(id, bucket)
	return bucket
}

func (m *BucketManager) get(id string) (*Bucket, bool) {
	m.bucketMux.RLock()
	bucket, ok := m.buckets[id]
	m.bucketMux.RUnlock()
	return bucket, ok
}

func (m *BucketManager) set(id string, bucket *Bucket) {
	m.bucketMux.Lock()
	m.buckets[id] = bucket
	m.bucketMux.Unlock()
}

func (m *BucketManager) delete(id string) {
	m.bucketMux.Lock()
	delete(m.buckets, id)
	m.bucketMux.Unlock()
}
