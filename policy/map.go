package policy

import (
	"fmt"
	"net/url"
)

const mapKeyword = "map"

type MapPolicy struct {
	opStringPolicy
	replacer *Replacer
}

const redirectKeyword = "redirect"

type RedirectPolicy struct {
	opStringPolicy
	replacer *Replacer
}

const opReplace = "replace"

func init() {
	regFactory(newOpStringPolicyFactory(mapKeyword, []string{opReplace}, "url", func(ops []string, val string) (Policy, error) {
		replacer, url, err := parseReplacerString(mapKeyword, ops, val)
		if err != nil {
			return nil, err
		}

		return &MapPolicy{opStringPolicy{mapKeyword, ops, url, func(ops []string, url string) string {
			if len(ops) > 0 {
				return "映射为 url 替换（" + url + "）后的内容"
			} else {
				return "映射 " + url + " 的内容"
			}
		}}, replacer}, nil
	}))

	regFactory(newOpStringPolicyFactory(redirectKeyword, []string{opReplace}, "url", func(ops []string, val string) (Policy, error) {
		replacer, url, err := parseReplacerString(redirectKeyword, ops, val)
		if err != nil {
			return nil, err
		}

		return &RedirectPolicy{opStringPolicy{redirectKeyword, ops, url, func(ops []string, url string) string {
			if len(ops) > 0 {
				return "302 跳转至 url 替换（" + url + "）后的地址"
			} else {
				return "302 跳转至 " + url
			}
		}}, replacer}, nil
	}))
}

func parseReplacerString(keyword string, ops []string, val string) (*Replacer, string, error) {
	if len(ops) == 0 {
		url, err := checkURL(val)
		return nil, url, err
	} else if len(ops) != 1 {
		return nil, "", fmt.Errorf(`%s too many ops: %v`, keyword, ops)
	} else if ops[0] != opReplace {
		return nil, "", fmt.Errorf(`%s unknown op: %s`, keyword, ops[0])
	} else {
		replacer, err := NewReplacer(val)
		return replacer, val, err
	}
}

func checkURL(rawurl string) (string, error) {
	_, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	} else {
		return rawurl, nil
	}
}

func (p *MapPolicy) URL(source string) string {
	if p.replacer == nil {
		return p.str
	} else {
		return p.replacer.Replace(source)
	}
}

func (p *RedirectPolicy) URL(source string) string {
	if p.replacer == nil {
		return p.str
	} else {
		return p.replacer.Replace(source)
	}
}
