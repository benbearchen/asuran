package net

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
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

func NewHttp(url string, r *http.Request, dial func(netw, addr string) (net.Conn, error)) (*HttpResponse, []byte, error) {
	client := &http.Client{Transport: &http.Transport{Dial: dial}, CheckRedirect: checkRedirect}

	var postBody []byte
	var body io.Reader = nil
	if r != nil && r.Method == "POST" {
		b, err := ioutil.ReadAll(r.Body)
		if err == nil {
			postBody = b
			body = bytes.NewReader(b)
		}
	}

	method := "GET"
	if r != nil {
		method = r.Method
	}

	req, err := http.NewRequest(method, url, body)
	if r != nil {
		req.Header = r.Header
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, postBody, err
	}

	return &HttpResponse{resp}, postBody, nil
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

func (r *HttpResponse) Header() http.Header {
	return r.resp.Header
}

func (r *HttpResponse) ResponseCode() int {
	return r.resp.StatusCode
}

func (r *HttpResponse) ProxyReturn(w http.ResponseWriter, wrap io.Writer) ([]byte, error) {
	defer r.resp.Body.Close()
	h := w.Header()
	for k, v := range r.Header() {
		h[k] = v
	}

	w.WriteHeader(r.ResponseCode())

	if wrap == nil {
		wrap = w
	}

	var b bytes.Buffer
	_, err := io.Copy(io.MultiWriter(&b, wrap), r.resp.Body)
	return b.Bytes(), err
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 0 {
		req.Header = via[0].Header
	}

	return nil
}
