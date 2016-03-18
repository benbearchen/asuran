package chaosjson

import (
	helper "github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/web/proxy/plugin/api"

	"fmt"
	"net/http"
)

func init() {
	api.RegisterHandler("chaosjson", "change json by Diablo", chaosjson)
}

func chaosjson(context *api.Context, targetURI string, w http.ResponseWriter, r *http.Request) {
	response, _, _, err := helper.NewHttp(targetURI, r, nil, false)
	if err != nil {
		w.WriteHeader(502)
		fmt.Fprintln(w, err)
		return
	}

	defer response.Close()

	bytes, err := response.ReadAllBytes()
	if err != nil {
		w.WriteHeader(502)
		fmt.Fprintln(w, err)
		return
	}

	bytes = bytes[:len(bytes)/2]
	w.Write(bytes)
}
