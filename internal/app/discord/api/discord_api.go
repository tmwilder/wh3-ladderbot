package api

import (
	"bytes"
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/commands"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Channel struct {
	ChannelId   string `json:"id"`
	ChannelName string `json:"name"`
}

type Message struct {
	MessageId string `json:"id"`
}

type MessageToPost struct {
	Content string `json:"content"`
}

const maxMessageCharsLength = 1900

func ReplaceChannelContents(guildId string, channelName string, contentLines []string) {
	// Select the channel matching our name
	foundChannel, channel := findChannel(channelName, guildId)
	if !foundChannel {
		log.Panicf("Unable to find channel: %s on guild: %s", channelName, guildId)
	}

	// Get all posts in the channel
	messages := GetMessages(channel.ChannelId)

	// Delete all posts in the channel
	if (len(messages)) > 0 {
		if len(messages) == 1 {
			// Evidently the discord bulk delete cannot handle exactly one msg and you must call the per-message API.
			deleteOneMessage(channel.ChannelId, messages[0].MessageId)
		} else {
			deleteOurMessagesInChannel(channel.ChannelId, messages)
		}
	}

	// Chunk up contents into N posts of max post length and post them
	postContentInChannel(channel.ChannelId, contentLines)
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
	return messages
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
	statusCode, _ := callDiscord(incrementalUrl, http.MethodPost, body)

	if statusCode != http.StatusNoContent {
		panic(fmt.Sprintf("Unable to delete messages - got non-204 code: %d", statusCode))
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

func postContentInChannel(channelId string, contentLines []string) (success bool) {
	content := ""
	for _, v := range contentLines {
		if len(content) >= maxMessageCharsLength {
			postOneMessage(channelId, content)
			content = ""
		} else {
			content += fmt.Sprintf("%s\n", v)
		}
	}
	postOneMessage(channelId, content)
	return true
}

func postOneMessage(channelId string, content string) (success bool) {
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
