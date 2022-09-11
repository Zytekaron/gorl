package gorl

import (
	"testing"
	"time"
)

func TestIntervalUtils(t *testing.T) {
	const unit = time.Second

	originExact := time.Now()                  // origin
	originAddHalf := originExact.Add(unit / 2) // origin + .5 sec
	reset1Exact := originExact.Add(unit)       // first reset
	reset1Half := originAddHalf.Add(unit)      // first reset + .5 sec
	reset2Exact := reset1Exact.Add(unit)       // second reset
	reset2Half := reset1Half.Add(unit)         // second reset + .5 sec
	reset3Exact := reset2Exact.Add(unit)       // third reset
	reset3Half := reset2Half.Add(unit)         // third reset + .5 sec

	count := intervalCount(originExact, reset3Exact, unit)
	if count != 3 {
		t.Error("expected intervals between originExact and reset3Exact to be 3, got", count)
	}
	count = intervalCount(originExact, reset3Half, unit)
	if count != 3 {
		t.Error("expected intervals between originExact and reset3Half to be 3, got", count)
	}

	//last := lastBefore(originExact, reset3Exact, unit)
	//if last != reset2Exact {
	//	t.Error("expected last time interval before reset3Exact to be reset2Exact")
	//}
	//last = lastBefore(originExact, reset3Half, unit)
	//if last != reset3Exact {
	//	t.Error("expected last time interval before reset3Half to be reset3Exact")
	//}

	next := nextAfter(originExact, reset1Half, unit)
	if next != reset2Exact {
		t.Error("expected next time interval after reset1Half to be reset2Exact")
	}
	next = nextAfter(originExact, reset2Exact, unit)
	if next != reset3Exact {
		t.Error("expected next time interval after reset2Exact to be reset3Exact")
	}
}
