package ops

import (
	"context"
	"flag"
	"html/template"
	"log"
	"fmt"
	"net/http"
	"path/filepath"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

var (
	staticDir = flag.String("static_path", "./static", "path to serve static assets from")
	templatesDir = flag.String("templates_path", "./templates", "Path to the folder of templates to serve")
	algodHost = flag.String("algod_host", "http://localhost:4001", "Host of algod client")
	algodTokenPath = flag.String("algod_token_path", "./priv/algod.token", "Path to algod token")
)


func RegisterRoutes(r *httprouter.Router, baseURL string) {

	fmt.Println("hello katya")

	r.ServeFiles("/static/*filepath", http.Dir(*staticDir))
	r.GET("/", handleIndex())
	r.GET("/health", handleHealth())
	r.GET("/account_info/:address", handleAccountInfo())
}

func handleIndex() httprouter.Handle {

	indexTmplPath := filepath.Join(*templatesDir, "index.html")
	indexTmpl, err := template.New("index").ParseFiles(indexTmplPath)
	if err != nil {
		log.Fatal("Error loading template `", indexTmplPath,"`:", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		data := struct {
			Current string
			Items []string
		}{
			Current: "item2",
			Items: []string{
				"item1",
				"item2",
				"item3",
			},
		}

		err = indexTmpl.Execute(w, data)
		if err != nil {
			log.Println("Failed to execute `templates/index.html`:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}
	}
}

func handleHealth() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("ok"))
	}
}

func handleAccountInfo() httprouter.Handle {

	token, err := os.ReadFile(*algodTokenPath)
	if err != nil {
		log.Fatal("error reading algod token from `", *algodTokenPath, "`:", err.Error())
	}

	algodClient, err := algod.MakeClient(*algodHost, string(token))
	if err != nil {
		log.Fatal("error making algod client:", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		addr := p.ByName("address")
		info, err := algodClient.AccountInformation(addr).Do(context.Background())
		if err != nil {
			log.Println("error getting account info for `", addr, ",:", err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte(fmt.Sprintf("%+v", info)))
	}
}

