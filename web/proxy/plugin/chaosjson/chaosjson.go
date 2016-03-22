package chaosjson

import (
	helper "github.com/benbearchen/asuran/net"
	"github.com/benbearchen/asuran/web/proxy/plugin/api"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var (
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func init() {
	api.RegisterHandler("chaosjson", "change json by Diablo", chaosjson)
}

func chaosjson(context *api.Context, targetURI string, w http.ResponseWriter, r *http.Request) {
	failHandler := func(statusCode int, err error) {
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, err)

		context.Log(statusCode, nil, nil, err)
	}

	response, _, _, err := helper.NewHttp(targetURI, r, nil, false)
	if err != nil {
		failHandler(502, err)
		return
	}

	defer response.Close()

	bytes, err := response.ReadAllBytes()
	if err == nil {
		bytes, err = dealJson(context.Policy.Setting(), bytes)
	}

	if err != nil {
		failHandler(502, err)
		return
	}

	w.Write(bytes)

	context.Log(200, nil, bytes, nil)
}

func dealJson(setting string, bytes []byte) ([]byte, error) {
	switch setting {
	case "none":
		return []byte{}, nil
	case "empty-object":
		return []byte("{}"), nil
	case "empty-array":
		return []byte("[]"), nil
	}

	var obj interface{}
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		return nil, err
	}

	switch setting {
	case "eat-any-one":
		obj = eatJson(obj, 1)
	case "change-any-one":
		obj = changeJson(obj, 1)
	}

	return json.Marshal(obj)
}

func eatJson(obj interface{}, times int) interface{} {
	switch obj := obj.(type) {
	case bool, float64, string:
		return nil
	case []interface{}:
		if len(obj) == 0 {
			return nil
		}

		i := r.Int() % len(obj)
		newly := make([]interface{}, i, len(obj))
		copy(newly, obj[0:i])
		return append(newly, obj[i+1:]...)
	case map[string]interface{}:
		if len(obj) == 0 {
			return nil
		}

		newly := make(map[string]interface{})
		i := r.Int() % len(obj)
		for k, v := range obj {
			if i != 0 {
				newly[k] = v
			}

			i--
		}

		return newly
	default:
		return nil
	}
}

func changeJson(obj interface{}, times int) interface{} {
	switch obj := obj.(type) {
	case bool:
		return !obj
	case float64:
		return obj + 1
	case string:
		if obj == "" {
			return "CCCCCCCC"
		} else {
			return ""
		}
	case []interface{}:
		if len(obj) == 0 {
			return []interface{}{0}
		}

		i := r.Int() % len(obj)
		newly := make([]interface{}, i, len(obj))
		copy(newly, obj[0:i])
		newly = append(newly, changeJson(obj[i], times-1))
		newly = append(newly, obj[i+1:]...)
		return newly
	case map[string]interface{}:
		if len(obj) == 0 {
			return make(map[string]interface{})
		}

		newly := make(map[string]interface{})
		i := r.Int() % len(obj)
		for k, v := range obj {
			if i == 0 {
				newly[k] = changeJson(v, times-1)
			} else {
				newly[k] = v
			}

			i--
		}

		return newly
	default:
		return nil
	}
}
