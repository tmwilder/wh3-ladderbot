package app

import (
	"bytes"
	"crypto/ed25519"
	"discordbot/internal/db"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
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
	// Authn - gets the user id
	// Do authz - checks that the userID can do the thing being attempted - bail w/4xx if not.
	switch interaction.Data.Name {
	case COMMAND_QUEUED:
		success := queueMatchRequest(interaction)
		if success {
			channelMessage = fmt.Sprintf("Queued request.")
		} else {
			channelMessage = fmt.Sprintf("Unable to queue request, please try again and contact admin if this persists.")
		}
		break
	case COMMAND_DEQUEUE:
		break
	default:
		panic("Unknown interaction: " + interaction.Data.Name)
	}
	return channelMessage
}

func queueMatchRequest(interaction Interaction) (success bool) {
	queueValue := interaction.Data.Options[0].Value
	discordUserId := interaction.Member.User.Id
	discordUserName := interaction.Member.User.Username

	// Get DB conn
	conn := db.GetGorm(db.GetMySQLConnStr())

	// Check to see if the user exists, if not create them.
	// We do this to avoid users ever having a register step - this takes advantage of Discord's Authn
	// and bot token validation flows.
	foundUser, user := db.GetUser(conn, discordUserId)
	if !foundUser {
		db.CreateUser(
			conn,
			db.User{
				DiscordId:       discordUserId,
				DiscordUserName: discordUserName,
				CurrentRating:   db.DEFAULT_RATING,
			})
		_, user = db.GetUser(conn, discordUserId)
	}

	// Now with assurances of a registered user - try to queue their entry
	didQueueMatch := db.CreateMatchRequest(conn, db.MatchRequest{
		RequestingUserId:  user.UserId,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RequestRange:      queueValue,       // TODO fix client param tooltip
		RequestedGameMode: db.GAME_MODE_ALL, // TODO add as client param
		MatchRequestState: db.MATCH_REQUEST_STATE_QUEUED,
	})

	// TODO run matchmaking algo here after we implement.
	db.FindPairing(conn, db.MatchRequest{})
	return didQueueMatch
}
