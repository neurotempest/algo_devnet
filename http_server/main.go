package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/neurotempest/algo_devnet/http_server/ops"
)

func main() {
	r := httprouter.New()
	ops.RegisterRoutes(r)

	listenAndServeForever(r, "localhost:1234")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-ch:
		log.Println("Received OS signal:", sig.String())
		return
	}
}

func listenAndServeForever(
	r *httprouter.Router,
	address string,
) {
	go func() {
		srv := http.Server{Addr: address, Handler: r}
		log.Println("Healthcheck listening at", address)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}
