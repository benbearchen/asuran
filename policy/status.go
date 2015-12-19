package policy

import (
	"fmt"
	"strconv"
)

type StatusPolicy struct {
	status int
}

const statusKeyword = "status"

func init() {
	regFactory(new(statusPolicyFactory))
}

type statusPolicyFactory struct {
}

func (*statusPolicyFactory) Keyword() string {
	return statusKeyword
}

func (*statusPolicyFactory) Build(args []string) (Policy, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf("status need a status code")
	}

	status, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, args, err
	}

	if status < 100 || status > 999 {
		return nil, args, fmt.Errorf("status %d out of range, should be 100-999", status)
	}

	return &StatusPolicy{status}, args[1:], nil
}

func (s *StatusPolicy) Keyword() string {
	return statusKeyword
}

func (s *StatusPolicy) Command() string {
	return statusKeyword + " " + strconv.Itoa(s.status)
}

func (s *StatusPolicy) Comment() string {
	return "状态码 " + strconv.Itoa(s.status)
}

func (s *StatusPolicy) Update(p Policy) error {
	if s.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keywrod: %s vs %s", s.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *StatusPolicy:
		s.status = p.status
	default:
		return fmt.Errorf("unmatch policy")
	}

	return nil
}

func (s *StatusPolicy) StatusCode() int {
	return s.status
}
