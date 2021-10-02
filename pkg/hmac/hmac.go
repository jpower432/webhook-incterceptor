package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"hash"
	"io/ioutil"
	"os"
	"strings"
)

type hmacValidator string
// validateMAC reports whether messageMAC is a valid HMAC tag for message.
func validMAC(message, messageMAC, key []byte, shaVersion string) (bool, error) {
	var mac hash.Hash

	switch shaVersion {
	case "sha256":
		mac = hmac.New(sha256.New, key)
	case "sha512":
		mac = hmac.New(sha512.New, key)
	default:
		return false, fmt.Errorf("unsupported SHA version: %s", shaVersion)
	}
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC), nil
}

// CreateHash a message with the key and return bytes.
func CreateHash(message, key []byte, shaVersion string) ([]byte, error) {

	var mac hash.Hash

	switch shaVersion {
	case "sha256":
		mac = hmac.New(sha256.New, key)
	case "sha512":
		mac = hmac.New(sha512.New, key)
	default:
		return nil, fmt.Errorf("unsupported SHA version: %s", shaVersion)
	}
	mac.Write(message)
	hash := mac.Sum(nil)
	return hash, nil
}

// Verify validates an encodedHash
func Verify(bytesIn []byte, encodedHash string, secretKey string, shaVersion string) error {
	var results error

	if strings.Contains(encodedHash, "=") {

		shaName := strings.Split(encodedHash, "=")
		if shaName[0] != shaVersion {
			return fmt.Errorf("incorrect hashing method: %s", shaName[0])
		}

		messageMAC := shaName[1]
		messageMACBuf, _ := hex.DecodeString(messageMAC)

		check, err := validMAC(bytesIn, []byte(messageMACBuf), []byte(secretKey), shaVersion)
		if err != nil {
			return err
		} else if check == false {
			results = fmt.Errorf("invalid message digest or key")
		}
	} else {
		results = fmt.Errorf("encoded hash does not contain a =")
	}

	return results
}

func (h hmacValidator) Intercept(ctx *gin.Context, results chan<- string) {

	sha := "sha256"
	header := os.Getenv("HEADER")
	logrus.Infof("Using header %s", header)
	signature := ctx.GetHeader(header)
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	secret := os.Getenv("WEBHOOK_SECRET")

	valid := Verify(body, signature, secret, sha)
	logrus.Info(valid)

	if valid == nil {
		logrus.Info("Sending results")
		results <- string(body)
	} else {
		results <- fmt.Sprintf("error: %v", valid)
		logrus.Error(valid)
	}
}

// Interceptor exported got plugin use
var Interceptor hmacValidator


