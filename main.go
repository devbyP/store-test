package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

var temps *template.Template
var db *pgx.Conn

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

	db = connectDB(os.Getenv("DB_CONNECT"))
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer migrateCancel()
	migrateProduct(migrateCtx)
	migrateCustomerInfo(migrateCtx)
	migrateProductTransactions(migrateCtx)

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
		Amount int64  `json:"amount"`
		Token  string `json:"token"`
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
		Amount:   payInfo.Amount,
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
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	TotalAmount int    `json:"totalAmount"`
}

// connect to database using connection string as parameter.
func connectDB(c string) *pgx.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	conn, err := pgx.Connect(ctx, c)
	if err != nil {
		log.Fatal("cannot connect to database")
	}
	// ping to recheck if the connection is ready.
	err = conn.Ping(ctx)
	if err != nil {
		log.Fatal("database connection ping fail")
	}
	// later assign to global variable
	return conn
}

// check and create table for products if not exist
// receive conntext from outside function.
// inject db object from outside function in case need multiple table for difference db.
// exit program if error.
func migrateProduct(ctx context.Context) {
	// price 100 = 1 thb
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS products(
		id uuid PRIMARY KEY,
		name varchar(50) NOT NULL,
		price integer default 20000
	)`)
	if err != nil {
		log.Fatal("error, create product table. " + err.Error())
	}
}

func migrateProductTransactions(ctx context.Context) {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS product_transactions(
			id uuid PRIMARY KEY,
			product_id uuid REFERENCE products (id) NOT NULL,
			customer_id uuid REFERENCE customer_info (id) NOT NULL,
			action varchar(10) DEFAULT 'out',
			out_amount integer NOT NULL,
			created_on timestamp NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		log.Fatal("error, create productTransactions table. " + err.Error())
	}
}

func migrateCustomerInfo(ctx context.Context) {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS customer_info(
		id uuid PRIMARY KEY,
		first_name varchar(100) NOT NULL,
		last_name varchar(100) NOT NULL,
		email varchar(155) NOT NULL,
		phone varchar(12)
	)`)
	if err != nil {
		log.Fatal("error , create customerInfo table. " + err.Error())
	}
}

// get the list of product from database.
func getDbProducts(ctx context.Context) ([]*Product, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, "SELECT id, name, price FROM products")
	if err != nil {
		return nil, err
	}
	prods := []*Product{}
	for rows.Next() {
		prod := &Product{}
		err := rows.Scan(&prod.ID, &prod.Name, &prod.Price)
		if err != nil {
			return nil, err
		}
		prods = append(prods, prod)
	}
	return prods, nil
}
