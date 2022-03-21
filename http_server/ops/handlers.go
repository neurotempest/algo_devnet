package ops

import (
	"errors"
	"flag"
	"html/template"
	"log"
	"fmt"
	"net/http"
	"path/filepath"
	"os"
	"crypto/rsa"
	"strings"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"

	"github.com/julienschmidt/httprouter"
	"github.com/golang-jwt/jwt"
)

var (
	staticDir = flag.String("static_path", "./static", "path to serve static assets from")
	templatesDir = flag.String("templates_path", "./templates", "Path to the folder of templates to serve")
	jwtPrivKeyPath = flag.String("jwt_priv", "./priv/private.key", "Path to rsa priv key for JWT signing")
	jwtPubKeyPath = flag.String("jwt_pub", "./priv/public.csr", "Path to rsa pub key for JWT signing")
)


func RegisterRoutes(r *httprouter.Router, baseURL string) {
	r.ServeFiles("/static/*filepath", http.Dir(*staticDir))
	r.GET("/", handleIndex())
	r.GET("/health", handleHealth())
	r.GET("/send_jwt", handleSendJWT(baseURL))
	r.POST("/receive_jwt", handleReceiveJWT())
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



func handleSendJWT(baseURL string) httprouter.Handle {

	jwtPrivKeyPem, err := os.ReadFile(*jwtPrivKeyPath)
	if err != nil {
		log.Fatal("error loading jwt priv key from `", *jwtPrivKeyPath, "`:", err.Error())
	}
	jwtPrivKey, err := jwt.ParseRSAPrivateKeyFromPEM(jwtPrivKeyPem)
	if err != nil {
		log.Fatal("error loading jwt priv key from PEM:", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		jwtToken, err := createJWTToken(jwtPrivKey, TokenData{
			Hello: "world",
		})
		if err != nil {
			log.Println("handleSendJWT: error creatign jwt token:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		req, err := http.NewRequest("POST", "http://localhost:1234/receive_jwt", strings.NewReader(jwtToken))
		if err != nil {
			log.Println("handleSendJWT: error making req:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		var c http.Client
		res, err := c.Do(req)
		if err != nil {
			log.Println("handleSendJWT: get error for receive_jwt:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if res.StatusCode != http.StatusOK {
			log.Println("handleSendJWT: get non 200 from receive_jwt:", res.StatusCode)
			http.Error(w, http.StatusText(res.StatusCode), res.StatusCode)
			return
		}
	}
}

func handleReceiveJWT() httprouter.Handle {

	jwtPubKeyBytes, err := os.ReadFile(*jwtPubKeyPath)
	if err != nil {
		log.Fatal("error loading jwt pub key from `", *jwtPubKeyPath, "`:", err.Error())
	}

	block, _ := pem.Decode(jwtPubKeyBytes)
	if block.Type != "CERTIFICATE REQUEST" {
		log.Fatal("unexpected pem type - expected CERTIFICATE REQUEST, got: ", block.Type)
	}

	log.Println("block.type:", block.Type)

	cert, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		log.Fatal("error parsing cert")
	}

	jwtPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		log.Fatal("error getting pub key from csr")
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("handleReceiveJWT: error reading req body:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		tokenData, err := decodeJWTToken(jwtPubKey, string(body))
		if err != nil {
			log.Println("handleReceiveJWT: error decoding token:", err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		log.Printf("Got token data: %+v\n", tokenData)
	}
}

type TokenData struct {
	Hello string
}

type Claims struct {
	jwt.StandardClaims
	TokenData
}

func createJWTToken(jwtPrivKey *rsa.PrivateKey, data TokenData) (string, error) {

	t := jwt.New(jwt.SigningMethodRS256)

	t.Claims = Claims{
		TokenData: data,
	}

	return t.SignedString(jwtPrivKey)
}

func decodeJWTToken(jwtPubKey *rsa.PublicKey, tokenString string) (TokenData, error) {


	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return TokenData{}, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtPubKey, nil
	})
	if err != nil {
		return TokenData{}, err
	}

	if !token.Valid {
		return TokenData{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return TokenData{}, errors.New("token.Claim conversion failed")
	}

	return claims.TokenData, nil
}
