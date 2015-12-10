package policy

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	delayKeyword   = "delay"
	timeoutKeyword = "timeout"
	dropKeyword    = "drop"
)

func init() {
	regFactory(&baseDelayPolicyFactory{
		delayKeyword,
		func() Policy { return new(DelayPolicy) },
	})

	regFactory(&baseDelayPolicyFactory{
		timeoutKeyword,
		func() Policy { return new(TimeoutPolicy) },
	})

	regFactory(&baseDelayPolicyFactory{
		dropKeyword,
		func() Policy { return new(DropPolicy) },
	})
}

type baseDelayPolicy struct {
	body     bool
	rand     bool
	duration float32
}

type DelayPolicy struct {
	baseDelayPolicy
}

type TimeoutPolicy struct {
	baseDelayPolicy
}

type DropPolicy struct {
	baseDelayPolicy
}

type baseDelayPolicyFactory struct {
	keyword string
	create  func() Policy
}

func (f *baseDelayPolicyFactory) Keyword() string {
	return f.keyword
}

func (f *baseDelayPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) < 1 {
		return nil, args, fmt.Errorf("`%d` need at least a duration")
	}

	body := false
	rand := false
	for len(args) > 0 {
		d := args[0]
		if d == "rand" {
			if rand {
				return nil, args, fmt.Errorf("too many `rand`")
			}

			rand = true
		} else if d == "body" {
			if body {
				return nil, args, fmt.Errorf("too many `body`")
			}

			body = true
		} else {
			break
		}

		args = args[1:]
	}

	if len(args) < 1 {
		return nil, args, fmt.Errorf("`%d` need a duration")
	}

	duration, err := parseDuration(args[0])
	if err != nil {
		return nil, args, err
	}

	args = args[1:]

	p := f.create()
	p.(iBaseDelayPolicy).init(body, rand, duration)
	return p, args, nil
}

type iBaseDelayPolicy interface {
	init(body, rand bool, duration float32)
}

func (p *baseDelayPolicy) init(body, rand bool, duration float32) {
	p.body = body
	p.rand = rand
	p.duration = duration
}

func (p *baseDelayPolicy) command() string {
	c := formatDuration(p.duration)
	if p.rand {
		c = "rand " + c
	}

	if p.body {
		c = "body " + c
	}

	return c
}

func (p *baseDelayPolicy) comment() string {
	c := formatDuration(p.duration)
	if p.rand {
		c = "随机（" + c + "）"
	} else {
		c = " " + c + " "
	}

	return c
}

func (p *baseDelayPolicy) Duration() time.Duration {
	return (time.Duration)(p.duration * 1000000000)
}

func (p *baseDelayPolicy) RandDuration(r *rand.Rand) time.Duration {
	t := p.duration
	if p.rand {
		t *= r.Float32()
	}

	return (time.Duration)(t * 1000000000)
}

func (p *DelayPolicy) Keyword() string {
	return delayKeyword
}

func (p *DelayPolicy) Command() string {
	return delayKeyword + " " + p.command()
}

func (p *DelayPolicy) Comment() string {
	c := "延时" + p.comment() + "后继续"
	if p.body {
		c = "对 HTTP Body "
	}

	return c
}

func (p *TimeoutPolicy) Keyword() string {
	return timeoutKeyword
}

func (p *TimeoutPolicy) Command() string {
	return timeoutKeyword + " " + p.command()
}

func (p *TimeoutPolicy) Comment() string {
	c := "等" + p.comment() + "后超时"
	if p.body {
		c = "对 HTTP Body "
	}

	return c
}

func (p *DropPolicy) Keyword() string {
	return dropKeyword
}

func (p *DropPolicy) Command() string {
	return dropKeyword + " " + p.command()
}

func (p *DropPolicy) Comment() string {
	c := "丢弃前" + p.comment() + "请求"
	if p.body {
		c = "对 HTTP Body "
	}

	return c
}

func parseDuration(d string) (float32, error) {
	var times float64 = 1
	if strings.HasSuffix(d, "ms") {
		d = d[:len(d)-2]
		times = 0.001
	} else if strings.HasSuffix(d, "h") {
		d = d[:len(d)-1]
		times = 60 * 60
	} else if strings.HasSuffix(d, "m") {
		d = d[:len(d)-1]
		times = 60
	} else if strings.HasSuffix(d, "s") {
		d = d[:len(d)-1]
	}

	f, err := strconv.ParseFloat(d, 32)
	if err != nil {
		return -1, err
	} else {
		return float32(f * float64(times)), nil
	}
}

func formatDuration(duration float32) string {
	if duration >= 60*60 {
		return strconv.FormatFloat(float64(duration/60/60), 'f', -1, 32) + "h"
	} else if duration >= 60 {
		return strconv.FormatFloat(float64(duration/60), 'f', -1, 32) + "m"
	} else if duration >= 1 {
		return strconv.FormatFloat(float64(duration), 'f', -1, 32) + "s"
	} else {
		return strconv.FormatFloat(float64(duration*1000), 'f', -1, 32) + "ms"
	}
}
