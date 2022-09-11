package gorl

import (
	"testing"
	"time"
)

func TestNewBucket(t *testing.T) {
	b := NewBucket(10, 25, time.Second)
	if b.Limit != 10 {
		t.Errorf("mismatched limit: expected '%d' but got '%d'", 10, b.Limit)
	}
	if b.Burst != 25 {
		t.Errorf("mismatched burst: expected '%d' but got '%d'", 25, b.Burst)
	}
	if b.Refill != time.Second {
		t.Errorf("mismatched refill: expected '%d' but got '%d'", time.Second, b.Refill)
	}
}

func TestBucketBasicTimings(t *testing.T) {
	tu1 := time.Now()
	tu2 := tu1.Add(time.Second)
	tu3 := tu2.Add(time.Second)
	b := NewBucket(5, 20, time.Second)

	// tu1: have=20 (init at burst cap)
	if !b.DrawAt(tu1, 15) {
		t.Error("expected to be able to draw 15 tokens at time unit 1")
	}
	// tu2: have=10 (5 from tu1, +5 from refill)
	if !b.DrawAt(tu2, 5) {
		t.Error("expected to be able to draw remaining 5 tokens at time unit 2 (have 10)")
	}
	// tu3: have=10 (5 from tu2, +5 from refill)
	if b.DrawAt(tu3, 11) {
		t.Error("expected to NOT be able to draw 11 tokens at time unit 2 (have 10)")
	}
	// tu3: have=10 (same as before)
	if !b.DrawAt(tu3, 10) {
		t.Error("expected to be able to draw 10 tokens at time unit 2 (have 10)")
	}
}

func TestBucketComplexTimings(t *testing.T) {
	tu1 := time.Now()
	tu2 := tu1.Add(time.Second)
	tu3 := tu2.Add(time.Second)
	tu4 := tu3.Add(time.Second)
	tu5 := tu4.Add(time.Second)
	tu6 := tu5.Add(time.Second)
	b := NewBucket(5, 20, time.Second)

	// tu1: have=20 (init at burst cap)
	if !b.DrawAt(tu1, 20) {
		t.Error("expected to be able to draw 15 tokens at time unit 1")
	}
	// skip tu2 (+5)
	// skip tu3 (+5)
	// tu4: have=15 (0 from tu1, +15 from refills including tu4)
	if !b.CanDrawAt(tu4, 15) {
		t.Error("expected to be able to draw 15 tokens at time unit 5")
	}
	if b.CanDrawAt(tu4, 20) {
		t.Error("expected to NOT be able to draw 20 tokens at time unit 5")
	}
	// skip tu5 (+5)
	// tu6: have=20 (15 from tu4, +10 from refills, capped at 20)
	if !b.CanDrawAt(tu6, 20) {
		t.Error("expected to be able to draw 20 tokens at time unit 6")
	}
	if b.CanDrawAt(tu6, 21) {
		t.Error("expected to NOT be able to draw 21 tokens at time unit 6")
	}
}

func TestBucket_Remaining(t *testing.T) {
	now := time.Now()
	b := NewBucket(10, 25, time.Second)

	b.ForceDrawAt(now, 10)
	remain := b.RemainingAt(now)
	if remain != 15 {
		t.Error("expected remaining tokens to be 15, got", remain)
	}
	b.ResetAt(now)

	b.ForceDrawAt(now, 25)
	remain = b.RemainingAt(now)
	if remain != 0 {
		t.Error("expected remaining tokens to be 0, got", remain)
	}
	b.ResetAt(now)

	b.ForceDrawAt(now, 50)
	remain = b.RemainingAt(now)
	if remain != 0 {
		t.Error("expected remaining tokens to be 0, got", remain)
	}
}

func TestBucket_Tokens(t *testing.T) {
	now := time.Now()
	b := NewBucket(10, 25, time.Second)

	b.ForceDrawAt(now, 10)
	tokens := b.TokensAt(now)
	if tokens != 15 {
		t.Error("expected token count to be 15, got", tokens)
	}
	b.ResetAt(now)

	b.ForceDrawAt(now, 25)
	tokens = b.TokensAt(now)
	if tokens != 0 {
		t.Error("expected token count to be 0, got", tokens)
	}
	b.ResetAt(now)

	b.ForceDrawAt(now, 50)
	tokens = b.TokensAt(now)
	if tokens != -25 {
		t.Error("expected token count to be -25, got", tokens)
	}
}

func TestBucket_Reset(t *testing.T) {
	now := time.Now()
	b := NewBucket(10, 25, time.Second)

	b.ForceDrawAt(now, 20)
	b.ResetAt(now)

	if b.Limit != 10 {
		t.Errorf("mismatched limit: expected '%d' but got '%d'", 10, b.Limit)
	}
	if b.Burst != 25 {
		t.Errorf("mismatched burst: expected '%d' but got '%d'", 25, b.Burst)
	}
	if b.Refill != time.Second {
		t.Errorf("mismatched refill: expected '%d' but got '%d'", time.Second, b.Refill)
	}

	b.ForceDrawAt(now, 20)
	b.ResetAt(now)

	if b.Limit != 10 {
		t.Errorf("mismatched limit: expected '%d' but got '%d'", 10, b.Limit)
	}
	if b.Burst != 25 {
		t.Errorf("mismatched burst: expected '%d' but got '%d'", 25, b.Burst)
	}
	if b.Refill != time.Second {
		t.Errorf("mismatched refill: expected '%d' but got '%d'", time.Second, b.Refill)
	}
}
