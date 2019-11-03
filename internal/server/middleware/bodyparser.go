package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)
const CtxBody = "ctx_body"

func ParseBody(next http.Handler, s interface{}, valid func(s interface{}) bool) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "application/json" {
			ctx := r.Context()
			decoder := json.NewDecoder(r.Body)
			if decoder.Decode(&s) != nil || !valid(&s) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.WithContext(context.WithValue(ctx, CtxBody, s))
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
