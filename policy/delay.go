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
		return nil, args, fmt.Errorf("`%s` need at least a duration", f.keyword)
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
		return nil, args, fmt.Errorf("`%s` need a duration", f.keyword)
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
		c = "随机 " + c + " "
	} else {
		c = " " + c + " "
	}

	return c
}

func (p *baseDelayPolicy) Body() bool {
	return p.body
}

func (p *baseDelayPolicy) Rand() bool {
	return p.rand
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

type baseDelayInterface interface {
	Body() bool
	Rand() bool
}

func (d *DelayPolicy) Keyword() string {
	return delayKeyword
}

func (d *DelayPolicy) Command() string {
	return delayKeyword + " " + d.command()
}

func (d *DelayPolicy) Comment() string {
	c := "延时" + d.comment() + "后继续"
	if d.body {
		c = "对 HTTP Body " + c
	}

	return c
}

func (d *DelayPolicy) Update(p Policy) error {
	if d.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keyword: %s vs %s", d.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *DelayPolicy:
		d.body = p.body
		d.rand = p.rand
		d.duration = p.duration
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (d *TimeoutPolicy) Keyword() string {
	return timeoutKeyword
}

func (d *TimeoutPolicy) Command() string {
	return timeoutKeyword + " " + d.command()
}

func (d *TimeoutPolicy) Comment() string {
	c := "等" + d.comment() + "后超时"
	if d.body {
		c = "等 HTTP Body 传输" + d.comment() + "后断开"
	}

	return c
}

func (d *TimeoutPolicy) Update(p Policy) error {
	if d.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keyword: %s vs %s", d.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *TimeoutPolicy:
		d.body = p.body
		d.rand = p.rand
		d.duration = p.duration
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (d *DropPolicy) Keyword() string {
	return dropKeyword
}

func (d *DropPolicy) Command() string {
	return dropKeyword + " " + d.command()
}

func (d *DropPolicy) Comment() string {
	c := "丢弃前" + d.comment() + "请求"
	if d.body {
		c = "对 HTTP Body "
	}

	return c
}

func (d *DropPolicy) Update(p Policy) error {
	if d.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keyword: %s vs %s", d.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *DropPolicy:
		d.body = p.body
		d.rand = p.rand
		d.duration = p.duration
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
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
