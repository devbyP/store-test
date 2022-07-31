// mock up storage for test.
package main

import "errors"

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
