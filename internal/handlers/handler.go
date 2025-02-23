package handlers

import "net/http"

type ImageProxyRequestHandler interface {
	Init()
	Handler(w http.ResponseWriter, r *http.Request)
}
