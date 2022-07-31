package main

import (
	"fmt"
	"strings"

	"github.com/omise/omise-go"
)

// keys should store somewhere else, like system environment variable, where other people cannot see them.
var (
	omisePublicKey  string
	omisePrivateKey string
)

const (
	privateKey = "s"
	publicKey  = "p"
)

// key validate should follow omise key text scheme.
func validKey(key, t string) bool {
	switch t {
	case privateKey:
		return strings.HasPrefix(key, "skey_")
	case publicKey:
		return strings.HasPrefix(key, "pkey_")
	default:
		return false
	}
}

// validate key before assign to the global variables.
func assignKey(public, private string) error {
	if !validKey(public, publicKey) {
		return fmt.Errorf("assignKey error, invalid public key\n%w ", omise.ErrInvalidKey)
	}
	if !validKey(private, privateKey) {
		return fmt.Errorf("assignKey error, invalid private(secret) key\n%w ", omise.ErrInvalidKey)
	}
	omisePublicKey, omisePrivateKey = public, private
	return nil
}
