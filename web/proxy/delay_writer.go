package proxy

import (
	gonet "github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/profile"

	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type delayWriter struct {
	delay      profile.DelayAction
	w          io.Writer
	d          time.Duration
	start      time.Time
	hasDelayed bool
	r          *rand.Rand
	first      bool
}

func newDelayWriter(delayAction profile.DelayAction, w io.Writer, r *rand.Rand) io.Writer {
	d := new(delayWriter)
	d.delay = delayAction
	d.w = w
	d.d = delayAction.RandDuration(r)
	d.start = time.Now()
	d.r = r
	d.first = true
	return d
}

func (d *delayWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return gonet.TryHijack(d.w)
}

func (d *delayWriter) Write(p []byte) (n int, err error) {
	if d.first {
		d.first = false
		d.Flush()
	}

	switch d.delay.Act {
	case profile.DelayActDelayEach:
		if !d.hasDelayed {
			d.hasDelayed = true
			<-time.NewTimer(d.d).C
		}
	case profile.DelayActTimeout:
		once := len(p) / 1024
		sum := 0
		if once < 1 {
			once = 1
		} else if once > 10*1024 {
			once = 10 * 1024
		}

		i := 0
		for {
			duration := time.Now().Sub(d.start)
			if duration >= d.d {
				gonet.ResetResponse(d.w)
				return sum, fmt.Errorf("http body timeout in %v(set %v)", duration, d.d)
			}

			c := len(p) - i
			if c == 0 {
				break
			} else if c > once {
				c = once
			}

			oc, err := d.w.Write(p[i : i+c])
			sum += oc
			i = sum
			if err != nil {
				return sum, err
			}
		}

		return sum, nil
	}

	return d.w.Write(p)
}

func (d *delayWriter) Flush() {
	if f, ok := d.w.(http.Flusher); ok {
		f.Flush()
	}
}
