package gorl

import (
	"sync"
	"time"
)

// Bucket is a thread-safe implementation of a leaky bucket.
// It allows a certain quantity of tokens to be drawn per given interval,
// and allows a burst of tokens to be drawn within a short period of time.
//
// Warning: If using the bucket with "At" methods, the times provided in
// each call (in order) must not chronologically descend. Using the non-At
// methods which use the current time is recommended for most use cases.
type Bucket struct {
	// Limit is the number of requests allowed per time unit, Refill.
	Limit int64
	// Burst is the number of requests allowed to be made at once.
	Burst int64
	// Refill is the interval at which Limit tokens are added back to
	// the bucket, with a maximum of Burst tokens.
	Refill time.Duration

	tokens     int64
	mux        sync.RWMutex
	lastUpdate time.Time
}

// NewBucket creates a new Bucket.
func NewBucket(limit, burst int64, refill time.Duration) *Bucket {
	return &Bucket{
		Limit:  limit,
		Burst:  burst,
		Refill: refill,
		tokens: burst,
	}
}

// CanDraw returns whether there are enough tokens remaining in the bucket to draw n.
func (b *Bucket) CanDraw(n int64) bool {
	return b.CanDrawAt(time.Now(), n)
}

// CanDrawAt returns whether there are enough tokens remaining in the bucket to draw n.
func (b *Bucket) CanDrawAt(t time.Time, n int64) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	return b.tokens >= n
}

// Draw draws n tokens from the bucket, returning whether there were enough tokens
// remaining to draw without overdraft. If not, no tokens are drawn from the bucket.
func (b *Bucket) Draw(n int64) bool {
	return b.DrawAt(time.Now(), n)
}

// DrawAt draws n tokens from the bucket, returning whether there were enough tokens
// remaining to draw without overdraft. If not, no tokens are drawn from the bucket.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens.
func (b *Bucket) DrawAt(t time.Time, n int64) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	if b.tokens < n {
		return false
	}
	b.tokens -= n
	return true
}

// DrawMax attempts to draw up to n tokens, returning the number of tokens drawn.
func (b *Bucket) DrawMax(n int64) int64 {
	return b.DrawMaxAt(time.Now(), n)
}

// DrawMaxAt attempts to draw up to n tokens, returning the number of tokens drawn.
func (b *Bucket) DrawMaxAt(t time.Time, n int64) int64 {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	drawn := min(n, b.tokens)
	b.tokens -= drawn
	return drawn
}

// ForceDraw forcefully draws a certain number of tokens and
// returns the number of remaining uses, which may be negative.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens.
// for potentially multiple refill intervals.
func (b *Bucket) ForceDraw(n int64) int64 {
	return b.ForceDrawAt(time.Now(), n)
}

// ForceDrawAt forcefully draws a certain number of tokens and
// returns the number of remaining uses, which may be negative.
//
// The number of tokens in the bucket increases as expected, so
// a large overdraft will result in a periodic absence of tokens
// for potentially multiple refill intervals.
func (b *Bucket) ForceDrawAt(t time.Time, n int64) int64 {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	b.tokens -= n
	return b.tokens
}

// SetTokens sets the number of available tokens and sets the last update time to the current time.
func (b *Bucket) SetTokens(tokens int64) {
	b.SetTokensAt(time.Now(), tokens)
}

// SetTokensAt sets the number of available tokens and sets the last update time to the provided time.
func (b *Bucket) SetTokensAt(t time.Time, tokens int64) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	b.tokens = tokens
}

// Remaining returns the remaining tokens which can be drawn.
//
// If the number of tokens in the bucket is less than zero, this returns 0.
func (b *Bucket) Remaining() int64 {
	return b.RemainingAt(time.Now())
}

// RemainingAt returns the remaining tokens which can be drawn at the specified time.
//
// If the number of tokens in the bucket is less than zero, this returns 0.
func (b *Bucket) RemainingAt(t time.Time) int64 {
	tokens := b.TokensAt(t)
	if tokens < 0 {
		return 0
	}
	return tokens
}

// Tokens returns the number of tokens in the bucket.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (b *Bucket) Tokens() int64 {
	return b.TokensAt(time.Now())
}

// TokensAt returns the number of tokens in the bucket at the specified time.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (b *Bucket) TokensAt(t time.Time) int64 {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.refill(t)

	return b.tokens
}

// InferTokensAt returns the number of tokens that will be in the bucket at the
// specified time. This assumes that there will be no modifications to the bucket
// between the current and provided time, such as Draw, Reset, or SetTokens.
//
// May be negative if tokens were overdrafted using SetTokens or ForceDraw.
func (b *Bucket) InferTokensAt(t time.Time) int64 {
	b.mux.RLock()
	defer b.mux.RUnlock()

	// determine how many times the refill interval will occur since the last update.
	delta := int64(t.Sub(b.lastUpdate) / b.Refill)

	// add the number of regenerated tokens to the current count
	tokens := b.tokens + delta*b.Limit
	if tokens > b.Burst {
		return b.Burst
	}
	return tokens
}

// NextRefill returns the next time this bucket will refill.
func (b *Bucket) NextRefill() time.Time {
	return b.NextRefillAt(time.Now())
}

// NextRefillAt returns the next time this bucket will refill, after the specified time.
func (b *Bucket) NextRefillAt(t time.Time) time.Time {
	b.mux.RLock()
	defer b.mux.RUnlock()
	b.refill(t)

	return nextAfter(b.lastUpdate, t, b.Refill)
}

// Reset resets this bucket. The number of tokens available is reset to
// the burst quantity, and the last update time is set to the current time.
func (b *Bucket) Reset() {
	b.ResetAt(time.Now())
}

// ResetAt resets this bucket. The number of tokens available is reset to
// the burst quantity, and the last update time is set to the provided time.
func (b *Bucket) ResetAt(t time.Time) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.tokens = b.Burst
	b.lastUpdate = t
}

// IsReset returns whether this bucket has just been created or is reset to
// a point where it can be fully drawn from up to the burst quantity.
func (b *Bucket) IsReset() bool {
	return b.IsResetAt(time.Now())
}

// IsResetAt returns whether this bucket has just been created or is reset to
// a point where it can be fully drawn from up to the burst quantity.
func (b *Bucket) IsResetAt(t time.Time) bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	b.refill(t)

	return b.tokens == b.Burst
}

// refill the tokens based on the last time it was updated and the current time.
//
// the bucket must be write-locked for the duration of the call.
func (b *Bucket) refill(t time.Time) {
	// if the bucket is already in a state where it is reset, change the lastUpdate time
	// to the current time to keep it in line with requests. this means a subsequent
	// request's refills will happen at the correct times, instead of being too early.
	if b.tokens == b.Burst {
		b.lastUpdate = t
		return // no need to check for refills
	}

	// determine how many times the refill interval has occurred since the last update.
	delta := intervalCount(b.lastUpdate, t, b.Refill)

	// skips `delta` time units, keeping lastUpdate in line with the initial time.
	b.skipDiff(delta)

	// add Limit tokens to the bucket for each Refill interval passed, capping at the burst
	// quantity. if the limit is exceeded, lastUpdate is reset to the current time to keep
	// it in line with requests (the same reason it resets at the top of this method).
	b.tokens += delta * b.Limit
	if b.tokens >= b.Burst {
		b.tokens = b.Burst
		b.lastUpdate = t
	}
}

// adds diff*refill to the lastUpdate (as opposed to just setting the lastUpdate
// to the current time. this ensures it always stays in line with the refill interval).
//
// the bucket must be write-locked for the duration of the call.
func (b *Bucket) skipDiff(diff int64) {
	mod := time.Duration(diff) * b.Refill
	b.lastUpdate = b.lastUpdate.Add(mod)
}
