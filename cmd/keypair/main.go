package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

func main() {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("could not generate key")
	}
	fmt.Printf("private key: %s\n", base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(pk)))
	fmt.Printf("public  key: %s\n", base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&pk.PublicKey)))
}
