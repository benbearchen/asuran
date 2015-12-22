package policy

import (
	"fmt"
)

type sub struct {
	keys map[string]string
}

func (s *sub) init(subKeys ...string) {
	s.keys = make(map[string]string)

	for _, k := range subKeys {
		s.reg(k)
	}
}

func (s *sub) reg(subKey string) {
	s.regFullKey(subKey, subKey)
}

func (s *sub) regFullKey(shortKey, fullKey string) {
	s.keys[shortKey] = fullKey
}

func (s *sub) isSubKey(keyword string) bool {
	_, ok := s.keys[keyword]
	return ok
}

func (s *sub) makeSub(keyword string, rest []string) (Policy, []string, error) {
	fullKey, ok := s.keys[keyword]
	if !ok {
		return nil, nil, fmt.Errorf(`invalid policy keyword: "%s", rest %v`, keyword, rest)
	}

	p, rest, err := keywordFactory(fullKey, rest)
	return p, rest, err
}
