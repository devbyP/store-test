package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

var temps *template.Template

// keys should store somewhere else, like system environment variable, where other people cannot see them.
var (
	omisePublicKey  string
	omisePrivateKey string
)

func main() {
	godotenv.Load()
	// initialize omise keys
	omisePublicKey, omisePrivateKey = os.Getenv("OmisePublicKey"), os.Getenv("OmisePrivateKey")
	if !strings.HasPrefix(omisePublicKey, "pkey_") || !strings.HasPrefix(omisePrivateKey, "skey_") {
		log.Fatal("missing omise key(s)")
	}

	// setting up server router using gorilla mux tools kit
	r := mux.NewRouter()
	temps = parseTemplate("./views/*.html")

	r.HandleFunc("/", serveStore)
	r.HandleFunc("/pay", handlePay)

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

// parsing views
// if any error exit the application
func parseTemplate(pattern string) *template.Template {
	temp, err := template.ParseGlob(pattern)
	if err != nil {
		log.Fatal(err)
	}
	// for later assign to global variable
	return temp
}

// dev version
// parse html template every time user this function. for easy refresh.
func serveStore(w http.ResponseWriter, r *http.Request) {
	storeTemp := template.Must(template.ParseFiles("./views/index.html"))
	storeTemp.Execute(w, nil)
}

// type use in handlePay function
type (
	// for request body
	// input object
	paymentInfoCreditCard struct {
		// sell item info
		ItemId         string `json:"itemId"`
		PurchaseAmount int    `json:"purchaseAmount"`

		// payment info
		// pay amount calculate at client.
		// need to recheck the pay amount from client and the item info again, before do the charge.
		PayAmount int64 `json:"payAmount"`

		// credit card encoded token receive from omise server.
		// client need to send credit card data directly to omise server.
		// after the client successfully get response from omise server(success status with token).
		// then client start request to this server with token and transaction info for server to handle.
		Token string `json:"token"`
	}

	// for response
	// output
	paymentSuccessResponse struct {
		Charge *omise.Charge            `json:"charge"`
		Create *operations.CreateCharge `json:"createCharge"`
	}
)

/*
  handler user credit card token and amount.
	function decode json object from body using json decoder method.
	create new client passing public key and private key from omise dashboard.
	check error incase wrong keys.
	omise api need to create charge and create charge struct before actual charge.
	after do the charge response back to user with charge and create charge.
	using json encoder encode the paymentSuccessResponse struct.
*/
func handlePay(w http.ResponseWriter, r *http.Request) {
	payInfo := paymentInfoCreditCard{}
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&payInfo)
	client, e := omise.NewClient(omisePublicKey, omisePrivateKey)
	if e != nil {
		log.Fatal(e)
	}
	charge, create := &omise.Charge{}, &operations.CreateCharge{
		Amount:   payInfo.PayAmount,
		Currency: "thb",
		Card:     payInfo.Token,
	}
	if e := client.Do(charge, create); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
	}
	jsonRes := &paymentSuccessResponse{Charge: charge, Create: create}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(jsonRes)
}

// get the products from database and send to client application.
func getProducts(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	prods, _ := prodCon.getDbProducts()
	encoder.Encode(prods)
}

var prodCon IProductDB = ProductController{}

var ProductNotFound = errors.New("product not found")

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	TotalAmount int     `json:"totalAmount"`
}

var products = map[string]*Product{
	"1234": {
		ID:          "1234",
		Name:        "Item1",
		Price:       1800,
		TotalAmount: 100,
	},
	"44fc": {
		ID:          "44fc",
		Name:        "Item2",
		Price:       2000,
		TotalAmount: 50,
	},
}

type IProductDB interface {
	getDbProduct(id string) *Product
	getDbProducts() (map[string]*Product, error)
	itemPurchaseUpdate(id string, amount int) error
}

type ProductController struct{}

func (p ProductController) getDbProduct(id string) *Product {
	return products[id]
}

func (p ProductController) itemPurchaseUpdate(id string, amount int) error {
	if _, ok := products[id]; !ok {
		return ProductNotFound
	}
	products[id].TotalAmount -= amount
	return nil
}

// get the list of product from database.
func (p ProductController) getDbProducts() (map[string]*Product, error) {
	return products, nil
}
