package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Unable to read standard input: %s", err.Error())
	}

	expectedVal := os.Getenv("Http_X-Hub-Signature")[5:] // first few chars are: sha1=
	fmt.Printf("Expected: %s\n", expectedVal)
	expectedBuf, _ := hex.DecodeString(expectedVal)

	secretKey := os.Getenv("secret_key")

	checked := CheckMAC(input, expectedBuf, []byte(secretKey))
	if checked == true {
		fmt.Println("The message was from your GitHub account.")
	} else {
		fmt.Println("The message was not from your GitHub account, or you don't share the same secret.")
	}
}

func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	fmt.Printf("MessageMAC: %x\n", messageMAC)
	fmt.Printf("CalculatedMAC: %x\n", expectedMAC)
	return hmac.Equal(messageMAC, expectedMAC)
}
