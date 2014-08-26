package proxy

import (
	"testing"
)

import (
	"github.com/benbearchen/asuran/profile"

	"bytes"
)

func TestSpeedWriter(t *testing.T) {
	w := newSpeedWriter(profile.SpeedAction{profile.SpeedActConstant, 10240}, new(bytes.Buffer)).(*speedWriter)
	if n, _ := w.next(10); n > 10 {
		t.Errorf("first next(10) return %d vs <=%d", n, 10)
	}

	if n, _ := w.next(1024); n > 1024 {
		t.Errorf("first next(1024) return %d vs <=%d", n, 1024)
	}

	if n, _ := w.next(10000); n > 10000 {
		t.Errorf("first next(10000) return %d vs <=%d", n, 10000)
	}

	if n, _ := w.next(100000); n > 10240 {
		t.Errorf("first next(100000) return %d vs <=%d", n, 10240)
	}

	w.wrote(5000)
	if n, w := w.next(5000); n > 5000 || w > 0 {
		t.Errorf("w5000 next(5000) return n(%d vs <=%d), w(%d vs %d", n, 5000, w, 0)
	}

	// time effect tests, OMG...
}
