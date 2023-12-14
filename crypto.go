package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func (node *Node) generateRSAKey(bits int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	node.privateKey = privateKey
	node.publicKey = &privateKey.PublicKey

	// Store private key in Node folder
	priDerText := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{
		Type: string(node.Address) + "-private Key",

		Headers: nil,

		Bytes: priDerText,
	}
	node_files_folder := "./tmp/" + node.Address
	privateHandler, err := os.Create(string(node_files_folder) + "/private.pem")
	if err != nil {
		panic(err)
	}
	defer privateHandler.Close()
	pem.Encode(privateHandler, &block)

	// Store public key in Node folder
	pubDerText, err := x509.MarshalPKIXPublicKey(node.publicKey)
	if err != nil {
		panic(err)
	}
	block = pem.Block{
		Type: string(node.Address) + "-public Key",

		Headers: nil,

		Bytes: pubDerText,
	}
	publicHandler, err := os.Create(string(node_files_folder) + "/public.pem")
	if err != nil {
		panic(err)
	}
	defer publicHandler.Close()
	pem.Encode(publicHandler, &block)

	id := hashString(string(node.Address))
	id.Mod(id, hashMod)
	idHandler, err := os.Create(string(node_files_folder) + "/" + id.String())
	if err != nil {
		panic(err)
	}
	defer idHandler.Close()

}

func (node *Node) encrypt(read_route string) ([]byte, error) {
	return rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		node.publicKey,
		ReadFileBytes(read_route),
		nil)

}

func (node *Node) decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		node.privateKey,
		ciphertext,
		nil)
}
