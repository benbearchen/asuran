package policy

import (
	"fmt"
	"strconv"
)

const nKeyword = "n"

func init() {
	regFactory(new(nPolicyFactory))
}

type NPolicy struct {
	n int
}

type nPolicyFactory struct {
}

func (*nPolicyFactory) Keyword() string {
	return nKeyword
}

func (*nPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf(`"n" need a number`)
	}

	n, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, args, err
	} else if n < 0 {
		return nil, args, fmt.Errorf(`"n %d" should >= 0`, n)
	}

	return &NPolicy{n}, args[1:], nil
}

func (n *NPolicy) Keyword() string {
	return nKeyword
}

func (n *NPolicy) Command() string {
	return nKeyword + " " + strconv.Itoa(n.n)
}

func (n *NPolicy) Comment() string {
	return "设定数量为 " + strconv.Itoa(n.n)
}

func (n *NPolicy) Update(p Policy) error {
	switch p := p.(type) {
	case *NPolicy:
		n.n = p.n
		return nil
	default:
		return fmt.Errorf("unmatch policy to NPolicy: %s", p.Command())
	}
}

func (n *NPolicy) N() int {
	return n.n
}
