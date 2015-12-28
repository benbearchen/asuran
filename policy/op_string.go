package policy

import (
	"fmt"
	"strings"
)

type opStringFactory struct {
	keyword string
	ops     []string
	argName string
	create  func([]string, string) (Policy, error)
}

func newOpStringPolicyFactory(keyword string, ops []string, argName string, create func([]string, string) (Policy, error)) *opStringFactory {
	if ops == nil {
		ops = []string{}
	}

	return &opStringFactory{keyword, ops, argName, create}
}

func (s *opStringFactory) Keyword() string {
	return s.keyword
}

func (s *opStringFactory) Build(args []string) (Policy, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf(`%d need optional ops(%v) and an arg "%s"`, s.keyword, s.ops, s.argName)
	}

	flags := make(map[string]bool)
	ops := make([]string, 0, len(s.ops))
	for len(args) > 0 {
		arg := args[0]
		found := false
		for _, op := range s.ops {
			if op == arg {
				found = true
				break
			}
		}

		if !found {
			break
		}

		_, ok := flags[arg]
		if ok {
			break // duplicate op??
		}

		flags[arg] = true
		ops = append(ops, arg)
		args = args[1:]
	}

	if len(args) == 0 {
		return nil, args, fmt.Errorf(`%d need a arg "%s"`, s.keyword, s.argName)
	}

	p, err := s.create(ops, args[0])
	if err != nil {
		return nil, args, err
	} else {
		return p, args[1:], nil
	}
}

type opStringPolicy struct {
	keyword string
	ops     []string
	str     string
	comment func([]string, string) string
}

func (s *opStringPolicy) Keyword() string {
	return s.keyword
}

func (s *opStringPolicy) Command() string {
	c := make([]string, 0, 2+len(s.ops))
	c = append(c, s.keyword)
	c = append(c, s.ops...)
	c = append(c, s.str)
	return strings.Join(c, " ")
}

func (s *opStringPolicy) Comment() string {
	return s.comment(s.ops, s.str)
}

func (s *opStringPolicy) Update(p Policy) error {
	if s.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keywrod: %s vs %s", s.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *opStringPolicy:
		s.ops = p.ops
		s.str = p.str
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (s *opStringPolicy) Op(op string) bool {
	for _, o := range s.ops {
		if op == o {
			return true
		}
	}

	return false
}
