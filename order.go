package main

import (
	"errors"
	"strconv"
)

var errOrderNotExists = errors.New("order not exists")

const (
	// status enum
	pending = iota
	paid
	cancel
	refund
)

type order struct {
	// product id as key
	// amount as value
	Purchase map[string]int
	Qty      int
	Status   int
	Owner    user
}

type orders map[string]*order

var orderStore orders

func (o orders) getOrder(id string) (*order, error) {
	var od *order
	var ok bool
	if od, ok = o[id]; !ok {
		return nil, errOrderNotExists
	}
	return od, nil
}

// return an id of new order
func (o orders) addOrder(od *order) string {
	id := incrementID()
	o[id] = od
	return id
}

func (o orders) statusUpdate(id string, s int) error {
	od, err := o.getOrder(id)
	if err != nil {
		return err
	}
	od.Status = s
	return nil
}

var idCount int

func incrementID() string {
	idCount++
	return strconv.Itoa(idCount)
}
