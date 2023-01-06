package discord

import (
	"bytes"
	"crypto/ed25519"
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/app/discord/interactions"
	"discordbot/internal/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func InteractionsHandler(c *gin.Context) {
	appConfig := config.GetAppConfig()

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
		switch interaction.Type {
		case 1:
			c.JSON(http.StatusOK, PingResponse{Type: 1})
			break
		case 2:
			message, shouldCrossPost := handleInteractionCommand(interaction)
			if shouldCrossPost {
				// Only announce things that are worth announcing.
				// Generally this means elo changes and queue changes do report and error states and operation
				// failures don't get reported.
				// TODO - enable when discord responds about the rate limits.
				//api.CrossPostMessageByName(interactions.LadderFeedChannel, message)
			}
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

func parseInteraction(requestBodyData []byte) (interaction interactions.Interaction) {
	err := json.Unmarshal(requestBodyData, &interaction)
	if err != nil {
		panic(err)
	}
	return interaction
}

func handleInteractionCommand(interaction interactions.Interaction) (channelMessage string, shouldCrossPost bool) {
	// Authn - gets the user id
	// Do authz - checks that the userID can do the thing being attempted - bail w/4xx if not.
	conn := db.GetDbConn()

	switch interaction.Data.Name {
	case commands.Queue:
		_, channelMessage, shouldCrossPost = interactions.Queue(conn, interaction)
		break
	case commands.Dequeue:
		_, channelMessage, shouldCrossPost = interactions.Dequeue(conn, interaction)
		break
	case commands.Report:
		_, channelMessage, shouldCrossPost = interactions.Report(conn, interaction)
	default:
		panic("Unknown interaction: " + interaction.Data.Name)
	}
	return channelMessage, shouldCrossPost
}
