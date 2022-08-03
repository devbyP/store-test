package main

import (
	"testing"
)

func TestAddOrder(t *testing.T) {
	testOrder := &order{}
	orderStore.addOrder(testOrder)
}
