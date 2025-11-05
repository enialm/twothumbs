// File: cmd/keygen/main.go

// This helper program prints a new activation code and API key.

package main

import (
	"fmt"
	"log"
	"twothumbs/internal/utils"
)

func main() {
	activationCode, err := utils.GenerateActivationCode()
	if err != nil {
		log.Fatal(err)
	}
	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Activation code: %s\nAPI key: %s\n", activationCode, apiKey)
}
