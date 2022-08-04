package main

import (
	"testing"
)

func TestAddOrder(t *testing.T) {
	testOrder := &order{}
	orderStore.addOrder(testOrder)
}

func TestGetOrder(t *testing.T) {
	testOrder := &order{}
	id := orderStore.addOrder(testOrder)
	order, err := orderStore.getOrder(id)
	if err != nil {
		t.Error(err)
	}
	t.Log(order)
}
