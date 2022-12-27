package app

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

const COMMAND_QUEUED = "queue"
const COMMAND_DEQUEUE = "dequeue"

type DiscordUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type DiscordMemberInfo struct {
	User DiscordUser `json:"user"`
}

type OptionData struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type InteractionData struct {
	Options []OptionData `json:"options"`
	Type    int          `json:"type"`
	Name    string       `json:"name"`
	Id      string       `json:"id"`
}

type Interaction struct {
	Type   int               `json:"type"`
	Token  string            `json:"token"`
	Member DiscordMemberInfo `json:"member"`
	Data   InteractionData   `json:"data"`
}

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

	requestBodyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	verified := verifyRequest(c, requestBodyData, key)
	if !verified {
		c.Status(http.StatusUnauthorized)
		c.Writer.Write([]byte("invalid request signature"))
	} else {
		interaction := parseInteraction(requestBodyData)
		fmt.Println(interaction)
		switch interaction.Type {
		case 1:
			c.JSON(http.StatusOK, PingResponse{Type: 1})
			break
		case 2:
			message := handleInteractionCommand(interaction)
			c.JSON(http.StatusOK, gin.H{"type": 4, "data": gin.H{"content": message}})
			break
		default:
			fmt.Println(interaction)
		}
	}
}

func verifyRequest(c *gin.Context, requestBodyData []byte, key ed25519.PublicKey) bool {
	var toVerify bytes.Buffer

	timestamp := c.Request.Header.Get("X-Signature-Timestamp")
	toVerify.WriteString(timestamp)

	sig := c.Request.Header.Get("X-Signature-Ed25519")

	byteSig, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	toVerify.Write(requestBodyData)
	return ed25519.Verify(key, toVerify.Bytes(), byteSig)
}

func parseInteraction(requestBodyData []byte) (interaction Interaction) {
	err := json.Unmarshal(requestBodyData, &interaction)
	if err != nil {
		panic(err)
	}
	return interaction
}

func handleInteractionCommand(interaction Interaction) (channelMessage string) {
	switch interaction.Data.Name {
	case COMMAND_QUEUED:
		// TODO do matchmaking stuff here later
		queueValue := interaction.Data.Options[0].Value
		channelMessage = fmt.Sprintf("Queued with ratings range: %d", queueValue)
		break
	case COMMAND_DEQUEUE:
		break
	default:
		panic("Unknown interaction: " + interaction.Data.Name)
	}
	return channelMessage
}
