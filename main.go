package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/slack-go/slack"
)

func main() {
	slackWebhookUrl := os.Getenv("SLACK_WEBHOOK_URL")

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	var found *types.Container
	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		if strings.Contains(container.Image, "minecraft-server") {
			c := container
			found = &c
			break
		}
	}
	if found == nil {
		log.Printf("not found minecraft-server container")
		os.Exit(1)
	}
	log.Printf("found container %s", found.ID)

	ctx := context.Background()
	playerNameToUUID := make(map[string]string)
	// collect users uuid
	{
		opt := types.ContainerLogsOptions{
			ShowStdout: true,
			Follow:     false,
			Tail:       "1000",
		}
		reader, err := cli.ContainerLogs(ctx, found.ID, opt)
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			row, err := ParseLog(scanner.Text()[11:])
			if err != nil {
				continue
			}
			switch r := row.(type) {
			case *PlayerUUID:
				playerNameToUUID[r.PlayerName] = r.UUID
			}
		}
		log.Printf("done collect users: %d", len(playerNameToUUID))
	}

	opt := types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
		Tail:       "0",
	}
	reader, err := cli.ContainerLogs(ctx, found.ID, opt)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		row := scanner.Text()
		// remove datetime example: [22:16:35] .....
		logMessage := row[11:]

		message := slack.WebhookMessage{
			IconEmoji: ":minecraft-glass:",
			Username:  "Minecraft Server",
		}

		res, _ := ParseLog(logMessage)
		switch r := res.(type) {
		case *ServerStarting:
			log.Printf("ServerStarting %s", r.Version)
			message.Text = fmt.Sprintf("Starting version:%s", r.Version)
		case *ServerStarted:
			log.Printf("ServerStarted %.3f sec", r.StartUp.Seconds())
			message.Text = fmt.Sprintf("Started %.1f sec", r.StartUp.Seconds())
		case *ServerStopping:
			log.Printf("ServerStopping")
			message.Text = fmt.Sprintf("Stop goodbye!")
		case *PlayerJoined:
			log.Printf("PlayerJoined %s", r.PlayerName)
			message.Username = r.PlayerName
			message.IconEmoji = ""
			message.IconURL = avatarUrl(playerNameToUUID[r.PlayerName])
			message.Text = fmt.Sprintf("_Joined_")
		case *PlayerLeft:
			log.Printf("PlayerLeft %s", r.PlayerName)
			message.Username = r.PlayerName
			message.IconEmoji = ""
			message.IconURL = avatarUrl(playerNameToUUID[r.PlayerName])
			message.Text = fmt.Sprintf("_Left_")
		case *ChatMessage:
			log.Printf("ChatMessage %s %s", r.PlayerName, r.Message)
			message.Username = r.PlayerName
			message.IconEmoji = ""
			message.IconURL = avatarUrl(playerNameToUUID[r.PlayerName])
			message.Text = r.Message
		case *PlayerUUID:
			log.Printf("PlayerUUID %s->%s", r.PlayerName, r.UUID)
			playerNameToUUID[r.PlayerName] = r.UUID
		default:
			// unknown log message
			// log.Printf("unknown log: '%s'", logMessage)
		}

		if message.Text != "" {
			log.Printf("post slack: %v", message)
			err = slack.PostWebhook(slackWebhookUrl, &message)
			if err != nil {
				log.Printf("post error %v", err)
			}
		}
	}
}

func avatarEmoji(username string) string {
	return fmt.Sprintf(":mc-%s:", username)
}

func avatarUrl(uuid string) string {
	return fmt.Sprintf("https://crafatar.com/avatars/%s", uuid)
}
