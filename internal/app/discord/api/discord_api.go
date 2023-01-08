package api

import (
	"bytes"
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/db"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Channel struct {
	ChannelId   string `json:"id"`
	ChannelName string `json:"name"`
}

type Message struct {
	MessageId string      `json:"id"`
	User      DiscordUser `json:"author"`
}

type MessageToPost struct {
	Content string `json:"content"`
}

type DiscordUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"bot"`
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
	Options []OptionData         `json:"options"`
	Type    int                  `json:"type"`
	Name    commands.CommandName `json:"name"`
	Id      string               `json:"id"`
}

type Interaction struct {
	Type   int               `json:"type"`
	Token  string            `json:"token"`
	Member DiscordMemberInfo `json:"member"`
	Data   InteractionData   `json:"data"`
}

type Role struct {
	RoleId   string `json:"id"`
	RoleName string `json:"name"`
}

const maxMessageCharsLength = 1800

func ReplaceChannelContents(guildId string, channelName string, contentLines []string) {
	// Select the channel matching our name
	foundChannel, channel := findChannel(channelName, guildId)
	if !foundChannel {
		log.Panicf("Unable to find channel: %s on guild: %s", channelName, guildId)
	}

	// Get all posts in the channel
	existingMessages := GetMessages(channel.ChannelId)
	posts := postFormatLines(contentLines)

	// Delete messages anyone else has posted.
	var deletedIndexes []int
	for i, v := range existingMessages {
		if v.User.IsBot == false {
			deleteOneMessage(channel.ChannelId, existingMessages[i].MessageId)
		}
	}
	for _, i := range deletedIndexes {
		existingMessages = append(existingMessages[:i], existingMessages[i+1:]...)
	}

	if len(existingMessages) > len(posts) {
		// If there's too many posts in the channel delete all posts and then post.
		if len(existingMessages) == 1 {
			// Evidently the discord bulk delete cannot handle exactly one msg and you must call the per-message API.
			deleteOneMessage(channel.ChannelId, existingMessages[0].MessageId)
		} else {
			deleteOurMessagesInChannel(channel.ChannelId, existingMessages)
		}
		postContentInChannel(channel.ChannelId, []Message{}, posts)
	} else {
		// Otherwise edit the existing messages to post and add new ones as needed.
		postContentInChannel(channel.ChannelId, existingMessages, posts)
	}

}

func findChannel(channelName string, guildId string) (foundChannel bool, channel Channel) {
	channels := GetChannels(guildId)
	for _, v := range channels {
		if v.ChannelName == channelName {
			return true, v
		}
	}
	return false, Channel{}
}

func GetChannels(guildId string) (channels []Channel) {
	incrementalUrl := fmt.Sprintf("guilds/%s/channels", guildId)
	statusCode, body := callDiscord(incrementalUrl, http.MethodGet, []byte{})

	if statusCode != http.StatusOK {
		panic(fmt.Sprintf("Unable to get channels - got non-200 code: %d", statusCode))
	}

	err := json.Unmarshal(body, &channels)

	if err != nil {
		log.Panicf("Unable to parse channel data: %v", err)
	}
	return channels
}

func GetMessages(channelId string) (messages []Message) {
	incrementalUrl := fmt.Sprintf("channels/%s/messages", channelId)
	statusCode, body := callDiscord(incrementalUrl, http.MethodGet, []byte{})

	if statusCode != http.StatusOK {
		panic(fmt.Sprintf("Unable to get messages - got non-200 code: %d", statusCode))
	}

	err := json.Unmarshal(body, &messages)

	if err != nil {
		log.Panicf("Unable to parse channel data: %v", err)
	}
	var output []Message
	for i := len(messages) - 1; i >= 0; i-- {
		output = append(output, messages[i])
	}
	return output
}

func deleteOurMessagesInChannel(channelId string, messages []Message) (success bool) {
	incrementalUrl := fmt.Sprintf("channels/%s/messages/bulk-delete", channelId)

	var messageIdList []string

	for _, v := range messages {
		messageIdList = append(messageIdList, v.MessageId)
	}

	postBody := map[string][]string{}

	postBody["messages"] = messageIdList

	body, err := json.Marshal(postBody)
	if err != nil {
		panic(err)
	}
	statusCode, body := callDiscord(incrementalUrl, http.MethodPost, body)

	if statusCode != http.StatusNoContent {
		panic(fmt.Sprintf("Unable to delete messages - got non-204 code: %d with response body: %s", statusCode, body))
	}
	return true
}

func deleteOneMessage(channelId string, messageId string) (success bool) {
	incrementalUrl := fmt.Sprintf("channels/%s/messages/%s", channelId, messageId)
	statusCode, _ := callDiscord(incrementalUrl, http.MethodDelete, []byte{})
	if statusCode != http.StatusNoContent {
		panic(fmt.Sprintf("Unable to delete messages - got non-204 code: %d", statusCode))
	}
	return true
}

func postContentInChannel(channelId string, existingMessages []Message, posts []string) (success bool) {
	for i, post := range posts {
		if i <= len(existingMessages)-1 {
			editOneMessage(channelId, existingMessages[i], post)
		} else {
			PostOneMessage(channelId, post)
		}
		time.Sleep(5 * time.Second)
	}
	return true
}

func postFormatLines(contentLines []string) (posts []string) {
	content := ""
	for _, v := range contentLines {
		if len(content) >= maxMessageCharsLength {
			posts = append(posts, content)
			content = ""
		} else {
			content += fmt.Sprintf("%s\n", v)
		}
	}
	posts = append(posts, content)
	return posts
}

func UpsertDmChannel(recipient db.User) (createdChannel bool, response Channel) {
	incrementalUrl := fmt.Sprintf("/users/@me/channels")
	postBody := map[string]string{"recipient_id": recipient.DiscordId}
	body, err := json.Marshal(postBody)
	if err != nil {
		panic(err)
	}
	statusCode, body := callDiscord(incrementalUrl, http.MethodPost, body)
	if statusCode != http.StatusOK {
		log.Printf("Unable to create DM channel - got non-200 code: %d w/msg: %s", statusCode, string(body))
		return false, Channel{}
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Unable to read message response for DM channel creation with user %d", recipient.UserId)
		return false, Channel{}
	}
	return true, response
}

func PostOneMessage(channelId string, content string) (success bool, response Message) {
	incrementalUrl := fmt.Sprintf("channels/%s/messages", channelId)
	postBody := map[string]string{"content": content}
	body, err := json.Marshal(postBody)
	if err != nil {
		panic(err)
	}
	statusCode, body := callDiscord(incrementalUrl, http.MethodPost, body)
	if statusCode != http.StatusOK {
		panic(fmt.Sprintf("Unable to post message - got non-204 code: %d w/msg: %s", statusCode, string(body)))
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Panicf("Unable to read message response for message posted to %s", channelId)
	}

	return true, response
}

func editOneMessage(channelId string, message Message, newContent string) (success bool, response Message) {
	incrementalUrl := fmt.Sprintf("channels/%s/messages/%s", channelId, message.MessageId)
	postBody := map[string]string{"content": newContent}
	body, err := json.Marshal(postBody)
	if err != nil {
		panic(err)
	}
	statusCode, body := callDiscord(incrementalUrl, http.MethodPatch, body)
	if statusCode != http.StatusOK {
		panic(fmt.Sprintf("Unable to update message - for message %s got non-200 code: %d w/msg: %s", message.MessageId, statusCode, string(body)))
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Panicf("Unable to read message response for message posted to %s", channelId)
	}

	return true, response
}

func CrossPostMessageByName(channelName string, message string) (success bool) {
	foundChannel, channel := findChannel(channelName, config.GetAppConfig().HomeGuildId)
	if !foundChannel {
		log.Panicf("Unable to find channel: %s", channelName)
		return false
	}
	posted, _ := PostOneMessage(channel.ChannelId, message)
	if !posted {
		log.Panicf("Unable to post message to channel: %s", channelName)
	}

	// TODO - enable when discord responds about the rate limits.
	//incrementalUrl := fmt.Sprintf("channels/%s/messages/%s/crosspost", channel.ChannelId, createdMessage.MessageId)
	//
	//status, crossPostRes := callDiscord(incrementalUrl, http.MethodPost, []byte{})
	//
	//if status != http.StatusOK {
	//	log.Panicf("Unable to cross post message: %s", crossPostRes)
	//}
	return true
}

func GetRoles(guildId string) (roles []Role) {
	incrementalUrl := fmt.Sprintf("/guilds/%s/roles", guildId)
	statusCode, body := callDiscord(incrementalUrl, http.MethodGet, []byte{})

	if statusCode != http.StatusOK {
		panic(fmt.Sprintf("Unable to get roles from guild - got non-200 code: %d", statusCode))
	}

	err := json.Unmarshal(body, &roles)

	if err != nil {
		log.Panicf("Unable to parse role data: %v", err)
	}

	return roles
}

func findRole(roleName string, guildId string) (foundRole bool, channel Role) {
	roles := GetRoles(guildId)
	for _, v := range roles {
		if v.RoleName == roleName {
			return true, v
		}
	}
	return false, Role{}
}

func AddRoleToGuildMember(roleName string, userId string) (success bool) {
	appConfig := config.GetAppConfig()
	foundRole, role := findRole(roleName, appConfig.HomeGuildId)

	if !foundRole {
		panic(fmt.Sprintf("Unable to find role with name: %s", roleName))
	}

	incrementalUrl := fmt.Sprintf("/guilds/%s/members/%s/roles/%s", appConfig.HomeGuildId, userId, role.RoleId)
	statusCode, _ := callDiscord(incrementalUrl, http.MethodPut, []byte{})

	if statusCode != http.StatusNoContent {
		panic(fmt.Sprintf("Unable to add role to member of the guild - got non-200 code: %d", statusCode))
	}

	return true
}

func RemoveRoleFromGuildMember(roleName string, userId string) (success bool) {
	appConfig := config.GetAppConfig()
	_, role := findRole(roleName, appConfig.HomeGuildId)

	incrementalUrl := fmt.Sprintf("/guilds/%s/members/%s/roles/%s", appConfig.HomeGuildId, userId, role.RoleId)
	statusCode, _ := callDiscord(incrementalUrl, http.MethodDelete, []byte{})

	if statusCode != http.StatusNoContent {
		panic(fmt.Sprintf("Unable to add role to member of the guild - got non-200 code: %d", statusCode))
	}

	return true
}

func callDiscord(incrementalUrl string, method string, serializedBody []byte) (statusCode int, body []byte) {
	url := fmt.Sprintf("%s/%s", commands.DiscordV10AppBase, incrementalUrl)

	client := &http.Client{}

	reader := bytes.NewReader(serializedBody)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		panic(err)
	}

	appConfig := config.GetAppConfig()
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", appConfig.DiscordBotToken))
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	body, readErr := ioutil.ReadAll(resp.Body)

	if err != nil || readErr != nil {
		panic("Failure to post command: " + fmt.Sprintf("%s", body))
	}
	statusCode = resp.StatusCode
	return statusCode, body
}
