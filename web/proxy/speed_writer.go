package proxy

import (
	gonet "github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/policy"

	"bufio"
	_ "fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type speedWriter struct {
	speed     *policy.SpeedPolicy
	w         io.Writer
	last      time.Time
	newBytes  int
	canSubPkg bool
}

func newSpeedWriter(speedAction *policy.SpeedPolicy, w io.Writer, canSubPackage bool) io.Writer {
	s := new(speedWriter)
	s.speed = speedAction
	s.w = w
	s.last = time.Time{}
	s.newBytes = 0
	s.canSubPkg = canSubPackage
	return s
}

func (t *speedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return gonet.TryHijack(t.w)
}

func (t *speedWriter) Write(p []byte) (n int, err error) {
	if !t.canSubPkg {
		return t.writeWait(p)
	}

	inputLen := len(p)
	for len(p) > 0 {
		once, w := t.next(len(p))
		if once > 0 {
			n, err = t.w.Write(p[:once])
			if err != nil {
				return
			}

			t.Flush()
			t.wrote(n)
			p = p[n:]
		} else {
			<-time.NewTimer(w).C
		}
	}

	return inputLen, nil
}

func (t *speedWriter) wrote(n int) {
	if t.last.IsZero() {
		t.last = time.Now()
	}

	t.newBytes += n
	//fmt.Printf("%15v wrote n: %6d   [bytes: %8d]\n", time.Now().Sub(t.last), n, t.newBytes)
}

func (t *speedWriter) next(c int) (int, time.Duration) {
	now := time.Now()
	if t.last.IsZero() {
		t.last = now
	}

	d := now.Sub(t.last)
	if d > time.Second {
		//fmt.Printf("%15v avg      %.2fKB/s\n", now.Sub(t.last), float64(int64(t.newBytes)*int64(time.Second)/int64(d))/1024)
	}

	maxLast := int(float32(d) * t.speed.Speed() / float32(time.Second))
	next := int(float32(d+time.Second/2)*t.speed.Speed()/float32(time.Second)) - maxLast
	if t.newBytes > 0 && t.newBytes < maxLast && d > 5*time.Second {
		t.last = now
		t.newBytes = 0
	} else if t.newBytes > maxLast {
		next -= t.newBytes - maxLast
	}

	if next <= 0 {
		//fmt.Printf("%15v sleep 0.5s\n", now.Sub(t.last))
		return 0, time.Second / 2
	}

	if next > c {
		next = c
	}

	//fmt.Printf("%15v next  n: %6d   [bytes: %8d]\n", now.Sub(t.last), next, t.newBytes)
	return next, 0
}

func (t *speedWriter) writeWait(p []byte) (int, error) {
	d := int(float32(len(p)) * 1000 / t.speed.Speed())
	//fmt.Println(len(p), "w", d)
	w := time.NewTimer(time.Duration(d) * time.Millisecond)
	n, err := t.w.Write(p)
	t.Flush()
	<-w.C
	return n, err
}

func (t *speedWriter) Flush() {
	if f, ok := t.w.(http.Flusher); ok {
		f.Flush()
	}
}
