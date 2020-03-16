package function

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
)

// Handle a serverless request
func Handle(req []byte) string {
	expectedVal := os.Getenv("Http_X-Hub-Signature")[5:] // first few chars are: sha1=
	fmt.Printf("Expected: %s\n", expectedVal)
	expectedBuf, _ := hex.DecodeString(expectedVal)

	secretKey := os.Getenv("secret_key")

	checked := CheckMAC(req, expectedBuf, []byte(secretKey))
	if checked == true {
		return fmt.Sprint("The message was from your GitHub account.")
	}
	return fmt.Sprint("The message was not from your GitHub account, or you don't share the same secret.")
}

func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	fmt.Printf("MessageMAC: %x\n", messageMAC)
	fmt.Printf("CalculatedMAC: %x\n", expectedMAC)
	return hmac.Equal(messageMAC, expectedMAC)
}
