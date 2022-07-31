package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	// initialize omise keys
	if err := assignKey(os.Getenv("OmisePublicKey"), os.Getenv("OmisePrivateKey")); err != nil {
		log.Fatalf("missing omise key(s), %s", err)
	}

	// setting up server router using gorilla mux tools kit
	r := mux.NewRouter()
	temps = parseTemplate("./views/*.html")

	r.HandleFunc("/", serveStore)
	r.HandleFunc("/pay", handleBuy)
	r.HandleFunc("/product", getProducts)

	// server config
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  20 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
