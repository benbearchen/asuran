package policy

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ChunkedOption int

const (
	ChunkedDefault ChunkedOption = iota
	ChunkedOn
	ChunkedOff
	ChunkedBlock
	ChunkedSize
)

const chunkedKeyword = "chunked"

type ChunkedPolicy struct {
	op     ChunkedOption
	blockN int
	sizes  []int
}

type chunkedPolicyFactory struct {
}

func init() {
	regFactory(new(chunkedPolicyFactory))
}

func (f *chunkedPolicyFactory) Keyword() string {
	return chunkedKeyword
}

func (f *chunkedPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) < 1 {
		return nil, args, fmt.Errorf(`chunked need a option: default|on|off|block...|size...`)
	}

	op := ChunkedDefault
	blockN := 0
	var sizes []int

	option, rest := args[0], args[1:]
	switch option {
	case "default":
	case "on":
		op = ChunkedOn
	case "off":
		op = ChunkedOff
	case "block":
		if len(rest) < 1 {
			return nil, rest, fmt.Errorf(`"chunked block" need a block count`)
		}

		n, err := strconv.Atoi(rest[0])
		if err != nil {
			return nil, rest, err
		} else if n <= 0 {
			return nil, rest, fmt.Errorf(``)
		}

		op = ChunkedBlock
		blockN = n
		rest = rest[1:]
	case "size":
		if len(rest) < 1 {
			return nil, rest, fmt.Errorf(`"chunked size" need a bytes serial for each chunked block, such as 1024,3,5`)
		}

		ss := strings.Split(rest[0], ",")
		sizes = make([]int, 0, len(ss))
		for _, s := range ss {
			if len(s) == 0 {
				continue
			}

			size, err := strconv.Atoi(s)
			if err != nil {
				return nil, rest, fmt.Errorf(`"chunked size" need a "integer" serial, not "%v"`, s)
			}

			sizes = append(sizes, size)
		}

		op = ChunkedSize
		rest = rest[1:]
	default:
		return nil, args, fmt.Errorf(`chunked need a option: default|on|off|block...|size..., not "%v"`, args[0])
	}

	return newChunkedPolicy(op, blockN, sizes), rest, nil
}

func newChunkedPolicy(option ChunkedOption, blockN int, sizes []int) *ChunkedPolicy {
	return &ChunkedPolicy{option, blockN, sizes}
}

func (c *ChunkedPolicy) Keyword() string {
	return chunkedKeyword
}

func (c *ChunkedPolicy) Command() string {
	cs := []string{chunkedKeyword}
	switch c.op {
	case ChunkedDefault:
		cs = append(cs, "default")
	case ChunkedOn:
		cs = append(cs, "on")
	case ChunkedOff:
		cs = append(cs, "off")
	case ChunkedBlock:
		cs = append(cs, "block")
		cs = append(cs, strconv.Itoa(c.blockN))
	case ChunkedSize:
		cs = append(cs, "size")
		sizes := make([]string, len(c.sizes))
		for i, b := range c.sizes {
			sizes[i] = strconv.Itoa(b)
		}

		cs = append(cs, strings.Join(sizes, ","))
	default:
		panic(fmt.Sprintf("invalid option: %v", c.op))
	}

	return strings.Join(cs, " ")
}

func (c *ChunkedPolicy) Comment() string {
	switch c.op {
	case ChunkedDefault:
		return "不干扰 Chunked 策略"
	case ChunkedOn:
		return "强制开启 Chunked"
	case ChunkedOff:
		return "强制关闭 Chunked"
	case ChunkedBlock:
		return "强制固定块数的 Chunked"
	case ChunkedSize:
		return "强制固定字节数的 Chunked 序列"
	default:
		return ""
	}
}

func (c *ChunkedPolicy) Update(p Policy) error {
	switch p := p.(type) {
	case *ChunkedPolicy:
		c.op = p.op
		c.blockN = p.blockN
		c.sizes = p.sizes
		return nil
	default:
		return fmt.Errorf(`unknown policy "%s": %v`, p.Keyword(), p)
	}
}

func (c *ChunkedPolicy) Option() ChunkedOption {
	return c.op
}

func (c *ChunkedPolicy) Block() int {
	if c.op != ChunkedBlock {
		return -1
	} else {
		return c.blockN
	}
}

func (c *ChunkedPolicy) Sizes() []int {
	if c.op != ChunkedSize {
		return nil
	} else {
		return c.sizes
	}
}

type ChunkedSizeQueue struct {
	sizes []int
	index int
}

func (q *ChunkedSizeQueue) Next() int {
	r := q.sizes[q.index]
	if q.index+1 < len(q.sizes) {
		q.index++
	}

	if r < 0 {
		return math.MaxInt32
	} else {
		return r
	}
}

func (c *ChunkedPolicy) SizesQueue() *ChunkedSizeQueue {
	sizes := c.Sizes()
	if len(sizes) > 0 {
		return &ChunkedSizeQueue{sizes, 0}
	}

	return nil
}
