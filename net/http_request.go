package net

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpResponse struct {
	resp *http.Response
}

func NewHttpGet(url string) (*HttpResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return &HttpResponse{resp}, nil
}

func (r *HttpResponse) Close() {
	if r.resp != nil {
		r.resp.Body.Close()
	}
}

func (r *HttpResponse) ReadAllBytes() ([]byte, error) {
	if r.resp != nil {
		return ioutil.ReadAll(r.resp.Body)
	}

	return make([]byte, 0), fmt.Errorf("invalid response")
}

func (r *HttpResponse) ReadAll() (string, error) {
	bytes, err := r.ReadAllBytes()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
