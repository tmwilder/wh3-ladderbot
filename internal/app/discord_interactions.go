package app

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

/*
	Standard response to Discord Ping request.
*/
type PingResponse struct {
	Type int `json:"type"`
}

func interactionsHandler(c *gin.Context) {
	appConfig := GetDiscordAppConfig()

	decodedKey, err := hex.DecodeString(appConfig.DiscordAppPublicKey)
	if err != nil {
		panic("Cannot decode key: " + appConfig.DiscordAppPublicKey)
	}
	key := ed25519.PublicKey(decodedKey)

	verified := verifyRequest(c, key)
	if !verified {
		c.Status(http.StatusUnauthorized)
		c.Writer.Write([]byte("invalid request signature"))
	} else {
		c.JSON(http.StatusOK, PingResponse{Type: 1})
	}
}

func verifyRequest(c *gin.Context, key ed25519.PublicKey) bool {
	var toVerify bytes.Buffer

	timestamp := c.Request.Header.Get("X-Signature-Timestamp")
	toVerify.WriteString(timestamp)

	sig := c.Request.Header.Get("X-Signature-Ed25519")

	byteSig, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		panic(err)
	}
	toVerify.Write(jsonData)
	return ed25519.Verify(key, toVerify.Bytes(), byteSig)
}
