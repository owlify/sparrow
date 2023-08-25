package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/tuvistavie/securerandom"
)

/*
Utility methods for http client calls
*/
func generateHmacSignature(serviceID, serviceKey string) (string, string) {
	nonce, _ := securerandom.Hex(16)
	key := fmt.Sprintf("%s-%s", nonce, serviceID)
	signature := hmac.New(sha1.New, []byte(serviceKey))
	signature.Write([]byte(key))
	serviceSignature := hex.EncodeToString(signature.Sum(nil))
	return nonce, serviceSignature
}

func headersForInternalRequest(serviceID, serviceKey string) Headers {
	nonce, serviceSignature := generateHmacSignature(serviceID, serviceKey)
	headers := Headers{
		"SIMPL-SERVICE-ID":        serviceID,
		"SIMPL-SERVICE-NONCE":     nonce,
		"SIMPL-SERVICE-SIGNATURE": serviceSignature,
	}
	return headers
}
