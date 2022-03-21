package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"flag"

	"github.com/julienschmidt/httprouter"
	"github.com/neurotempest/algo_devnet/http_server/ops"
)

var (
	baseURL = flag.String("url", "localhost:1234", "Base URL for http server")
)

func main() {
	flag.Parse()

	r := httprouter.New()
	ops.RegisterRoutes(r, *baseURL)

	listenAndServeForever(r, *baseURL)

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
		log.Println("Hello fgsda")
		log.Println("Healthcheck listening at", address)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}
