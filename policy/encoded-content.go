package policy

import (
	"net/url"
)

func checkEncodedContent(content string) error {
	_, err := url.QueryUnescape(content)
	return err
}

func decodeContent(content string) ([]byte, error) {
	r, err := url.QueryUnescape(content)
	if err != nil {
		return nil, err
	} else {
		return []byte(r), nil
	}
}
