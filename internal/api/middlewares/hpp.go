package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	WhiteList                   []string
}

func Hpp(options HPPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, options.CheckBodyOnlyForContentType) {
				// To filter the body params
				filterBodyParams(r, options.WhiteList)
			}

			if options.CheckQuery && r.URL.Query() != nil {
				// To filter the query params
				filterQueryParams(r, options.WhiteList)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("filterBodyParams", err)
		return
	}

	for k, v := range r.Form {
		if len(v) > 1 {
			// r.Form.Set(k, v[0]) // for first value
			r.Form.Set(k, v[len(v)-1]) // for last value
		}

		if isWhiteListed(k, whitelist) {
			// if not
			delete(r.Form, k)
		}
	}

}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()

	for k, v := range query {
		if len(v) > 1 {
			query.Set(k, v[0]) // for first value
			// query.Set(k, v[len(v)-1]) // for last value
		}

		if isWhiteListed(k, whitelist) {
			// if not
			query.Del(k)
		}
	}

	r.URL.RawQuery = query.Encode()

}

// To discard all the keys that do not match with the keys in the whitelist
func isWhiteListed(param string, whitelist []string) bool {
	for _, v := range whitelist {
		if param == v {
			return true
		}
	}

	return false

}
