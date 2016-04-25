package policy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const requestHeadersKeyword = "request-headers"
const responseHeadersKeyword = "response-headers"

type headerSettingAction int

const (
	HSA_ADD headerSettingAction = iota
	HSA_DELETE
	HSA_MODIFY
)

type headerSetting struct {
	act   headerSettingAction
	key   string
	value *string
}

type HeadersPolicy struct {
	stringPolicy
	headers []headerSetting
}

func init() {
	reg := func(keyword, comment string) {
		regFactory(newStringPolicyFactory(keyword, "header-settings", func(val string) (Policy, error) {
			headers, err := parseHeaderSettings(val)
			if err != nil {
				return nil, err
			}

			return &HeadersPolicy{stringPolicy{keyword, val, func(val string) string {
				return comment
			}}, headers}, nil
		}))
	}

	reg(requestHeadersKeyword, "设定请求 HTTP Headers")
	reg(responseHeadersKeyword, "设定回复 HTTP Headers")
}

func (h *HeadersPolicy) Apply(headers http.Header) {
	for _, s := range h.headers {
		switch s.act {
		case HSA_ADD:
			v, ok := headers[s.key]
			if ok && v != nil {
				headers[s.key] = append(v, *s.value)
			} else {
				headers[s.key] = []string{*s.value}
			}
		case HSA_DELETE:
			if s.value == nil {
				headers[s.key] = nil
			} else if v, ok := headers[s.key]; ok {
				a := make([]string, 0, len(v))
				if len(v) > 0 {
					for _, v := range v {
						if v == *s.value {
							continue
						}

						a = append(a, v)
					}
				}

				if len(a) > 0 {
					headers[s.key] = a
				} else {
					headers[s.key] = nil
				}
			}
		case HSA_MODIFY:
			headers[s.key] = []string{*s.value}
		default:
		}
	}
}

func (h *HeadersPolicy) Update(p Policy) error {
	if h.Keyword() != p.Keyword() {
		return fmt.Errorf("unmatch keyword: %s vs %s", h.Keyword(), p.Keyword())
	}

	switch p := p.(type) {
	case *HeadersPolicy:
		h.str = p.str
		h.headers = p.headers
	default:
		return fmt.Errorf("unmatch policy: %v", p)
	}

	return nil
}

func parseHeaderSettings(setting string) ([]headerSetting, error) {
	uss, err := url.QueryUnescape(setting)
	if err != nil {
		return nil, err
	}

	ss := strings.Split(uss, "\n")
	settings := make([]headerSetting, 0, len(ss))
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}

		act := HSA_MODIFY
		switch s[0] {
		case '+':
			act = HSA_ADD
			s = s[1:]
		case '-':
			act = HSA_DELETE
			s = s[1:]
		case '=':
			act = HSA_MODIFY
			s = s[1:]
		default:
		}

		s = strings.TrimSpace(s)
		kv := strings.SplitN(s, ":", 2)
		if len(kv) == 0 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		if len(key) == 0 {
			return nil, fmt.Errorf("header's key must NOT be empty: %q", s)
		}

		var value *string
		if len(kv) > 1 {
			value = &kv[1]
		}

		if act != HSA_DELETE && value == nil {
			return nil, fmt.Errorf("header modify NEED a value: %q", s)
		}

		settings = append(settings, headerSetting{act, key, value})
	}

	return settings, nil
}
