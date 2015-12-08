package policy

import (
	cmdarg "github.com/benbearchen/asuran/util/cmd"

	"fmt"
	"sync"
)

func Factory(cmd string) (Policy, error) {
	args := cmdarg.SplitCommand(cmd)
	p, _, err := innerFactory(args)
	return p, err
}

func innerFactory(args []string) (Policy, []string, error) {
	factoryMutex.RLock()
	defer factoryMutex.RUnlock()

	if len(args) == 0 {
		return nil, args, fmt.Errorf("%s", "empty command")
	}

	keyword, rest := args[0], args[1:]
	return keywordFactory(keyword, rest)
}

func keywordFactory(keyword string, rest []string) (Policy, []string, error) {
	f, ok := factory[keyword]
	if ok {
		return f.Build(rest)
	}

	return nil, rest, fmt.Errorf("unknown keyword: %s, rest: %v", keyword, rest)
}

type policyFactory interface {
	Keyword() string
	Build(args []string) (policy Policy, rest []string, err error)
}

func regFactory(f policyFactory) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()

	factory[f.Keyword()] = f
}

var factoryMutex sync.RWMutex
var factory map[string]policyFactory = make(map[string]policyFactory)
