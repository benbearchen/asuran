package policy

import (
	"net/url"
)

const mapKeyword = "map"

type MapPolicy struct {
	stringPolicy
}

const redirectKeyword = "redirect"

type RedirectPolicy struct {
	stringPolicy
}

func init() {
	regFactory(newStringPolicyFactory(mapKeyword, "url", func(url string) (Policy, error) {
		url, err := checkURL(url)
		if err != nil {
			return nil, err
		} else {
			return &MapPolicy{stringPolicy{mapKeyword, url, func(url string) string {
				return "映射 " + url + " 的内容"
			}}}, nil
		}
	}))

	regFactory(newStringPolicyFactory(redirectKeyword, "url", func(url string) (Policy, error) {
		url, err := checkURL(url)
		if err != nil {
			return nil, err
		} else {
			return &MapPolicy{stringPolicy{redirectKeyword, url, func(url string) string {
				return "302 跳转至 " + url
			}}}, nil
		}
	}))
}

func checkURL(rawurl string) (string, error) {
	_, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	} else {
		return rawurl, nil
	}
}
