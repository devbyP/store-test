package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

// dev version
// parse html template every time user this function. for easy refresh.
func serveStore(w http.ResponseWriter, r *http.Request) {
	storeTemp := template.Must(template.ParseFiles("./views/store.html"))
	prods, _ := prodCon.getDbProducts()
	prodsList := make([]*Product, len(prods))
	i := 0
	for _, prod := range prods {
		prodsList[i] = prod
		i++
	}
	storeTemp.Execute(w, map[string][]*Product{"products": prodsList})
}

// payment page.
func servePayment(w http.ResponseWriter, r *http.Request) {
	orderId := r.URL.Query().Get("order")
	paymentTemp := template.Must(template.ParseFiles("./views/payment.html"))
	order, err := orderStore.getOrder(orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if order.Status != pending {
		http.Error(w, "order already closed", http.StatusBadRequest)
	}
	paymentTemp.Execute(w, nil)
}

type purchaseItem struct {
	ID   string
	Name string
}

type ownershipPageData struct {
	PurchaseItems []*purchaseItem
	Customer      *customer
	OrderId       string
}

func serveOwnershipInput(w http.ResponseWriter, r *http.Request) {
	osTemp := template.Must(template.ParseFiles("./views/ownership.html"))
	orderId := r.URL.Query().Get("order")
	order, err := orderStore.getOrder(orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// prepare for the owner input list on client.
	var itemList []*purchaseItem
	pc := &ProductController{}
	for id, qty := range order.Purchase {
		prod := pc.getDbProduct(id)
		for i := 0; i < qty; i++ {
			// send item name and id to client.
			pi := &purchaseItem{ID: id, Name: prod.Name}
			itemList = append(itemList, pi)
		}
	}
	// response data
	osData := ownershipPageData{
		PurchaseItems: itemList,
		Customer:      order.Owner,
		OrderId:       orderId,
	}
	osTemp.Execute(w, osData)
}

// each item customer buy.
type customerPurchase struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

type customerOrder struct {
	Cart     []customerPurchase `json:"cart"`
	Customer customer           `json:"customer"`
}

func processOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cusOrder customerOrder
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&cusOrder)
	if err != nil {
		log.Fatal(err)
	}
	if len(cusOrder.Cart) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\":\"cannot process with empty cart\"}"))
		return
	}
	// register order to order store
	systemOrder := &order{
		Status:   pending,
		Owner:    &cusOrder.Customer,
		Purchase: map[string]int{},
	}
	for _, item := range cusOrder.Cart {
		systemOrder.Purchase[item.ID] = item.Qty
	}
	// get order id
	orderId := orderStore.addOrder(systemOrder)
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	encoder.Encode(map[string]string{"orderId": orderId})
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
func handleBuy(w http.ResponseWriter, r *http.Request) {
	payInfo := paymentInfoCreditCard{}
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&payInfo)
	charge, err := ChargeCard(payInfo.Token, payInfo.PayAmount)
	if errors.Unwrap(err) == omise.ErrInvalidKey {
		log.Fatal(err)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(charge)
}

// seperate logic for readability
// fix charge operation config to default.
// required amount and card info.
func ChargeCard(token string, amount int64) (*omise.Charge, error) {
	client, e := omise.NewClient(omisePublicKey, omisePrivateKey)
	if e != nil {
		return nil, fmt.Errorf("error create omise client in ChargeCard: %w", e)
	}
	charge, create := &omise.Charge{}, &operations.CreateCharge{
		Amount:      amount,
		Currency:    "thb",
		Card:        token,
		Description: "test charge description",
	}
	if e := client.Do(charge, create); e != nil {
		return nil, fmt.Errorf("error charge process ChargeCard: %w", e)
	}
	return charge, nil
}

// get the products from database and send to client application.
func getProducts(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	prods, _ := prodCon.getDbProducts()
	w.Header().Set("Content-type", "appication/json")
	encoder.Encode(prods)
}
