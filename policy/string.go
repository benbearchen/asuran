package policy

import (
	"fmt"
)

type stringPolicyFactory struct {
	keyword string
	argName string
	create  func(string) (Policy, error)
}

func newStringPolicyFactory(keyword, argName string, create func(string) (Policy, error)) *stringPolicyFactory {
	return &stringPolicyFactory{keyword, argName, create}
}

func (s *stringPolicyFactory) Keyword() string {
	return s.keyword
}

func (s *stringPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf(`%s need an arg "%s"`, s.keyword, s.argName)
	}

	p, err := s.create(args[0])
	if err != nil {
		return nil, args, err
	} else {
		return p, args[1:], nil
	}
}

type stringPolicy struct {
	keyword string
	str     string
	comment func(string) string
}

func (s *stringPolicy) Keyword() string {
	return s.keyword
}

func (s *stringPolicy) Command() string {
	return s.keyword + " " + s.str
}

func (s *stringPolicy) Comment() string {
	return s.comment(s.str)
}

func (s *stringPolicy) Update(p Policy) error {
	if s.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keywrod: %s vs %s", s.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *stringPolicy:
		s.str = p.Value()
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (s *stringPolicy) Value() string {
	return s.str
}
