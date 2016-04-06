package policy

import (
	"fmt"
)

const removeKeyword = "remove"

type RemovePolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(removeKeyword, "url's keyword", func(keyword string) (Policy, error) {
		if !urlSubKeys.isSubKey(keyword) {
			return nil, fmt.Errorf(`"url remove" need a url's keyword`)
		} else {
			return &RemovePolicy{stringPolicy{removeKeyword, keyword, func(keyword string) string {
				return "移除 " + keyword + " 子策略"
			}}}, nil
		}
	}))
}
