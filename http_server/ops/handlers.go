package ops

import (
	"github.com/julienschmidt/httprouter"
)

func RegisterRoutes(r *httprouter.Router) {
	r.GET("/health", handleHealth())
}

func handleHealth() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("ok"))
	}
}

