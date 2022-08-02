package main

type user struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	ReceiveNews bool   `json:"receiveNews"`
}
