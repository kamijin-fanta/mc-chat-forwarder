package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ErrUnknownLog = fmt.Errorf("unknown log")

var (
	chatRegex = regexp.MustCompile(`^\[Async Chat Thread - #\d+\/INFO\]: <(\w*)> (.+)$`)
	uuidRegex = regexp.MustCompile(`^\[User Authenticator #\d+\/INFO\]: UUID of player (\w*) is ([\w-]*)$`)
)

func ParseLog(log string) (MessageVariant, error) {
	if strings.HasPrefix(log, "[Server thread/INFO]: Starting minecraft server version ") {
		return &ServerStarting{
			Version: strings.TrimPrefix(log, "[Server thread/INFO]: Starting minecraft server version "),
		}, nil
	}
	if strings.HasPrefix(log, "[Server thread/INFO]: Done (") {
		var sec int
		var msec int
		fmt.Sscanf(log, "[Server thread/INFO]: Done (%d.%03ds)!", &sec, &msec)
		return &ServerStarted{
			StartUp: time.Duration(sec)*time.Second + time.Duration(msec)*time.Millisecond,
		}, nil
	}
	if strings.HasPrefix(log, "[Server thread/INFO]: Stopping server") {
		return &ServerStopping{}, nil
	}
	if strings.HasSuffix(log, "joined the game") {
		var playerName string
		fmt.Sscanf(log, "[Server thread/INFO]: %s joined the game", &playerName)
		return &PlayerJoined{
			PlayerName: playerName,
		}, nil
	}
	if strings.HasSuffix(log, "left the game") {
		var playerName string
		fmt.Sscanf(log, "[Server thread/INFO]: %s left the game", &playerName)
		return &PlayerLeft{
			PlayerName: playerName,
		}, nil
	}
	if strings.HasPrefix(log, "[Async Chat Thread") {
		matches := chatRegex.FindStringSubmatch(log)
		if len(matches) == 3 {
			return &ChatMessage{
				PlayerName: matches[1],
				Message:    matches[2],
			}, nil
		}
	}
	// uuid
	// [User Authenticator #26/INFO]: UUID of player kamijin is 98452fb7-def2-4253-8e4a-0ae57aa02604
	if strings.HasPrefix(log, "[User Authenticator #") {
		matches := uuidRegex.FindStringSubmatch(log)
		if len(matches) == 3 {
			return &PlayerUUID{
				PlayerName: matches[1],
				UUID:       matches[2],
			}, nil
		}
	}

	return nil, fmt.Errorf("unknown log: %s %w", log, ErrUnknownLog)
}

type MessageVariant interface {
	Type() string
}

// [Server thread/INFO]: Starting minecraft server version 1.19.2
type ServerStarting struct {
	Version string
}

func (s *ServerStarting) Type() string {
	return "ServerStarting"
}

// [Server thread/INFO]: Done (00.000s)! For help, type "help"
type ServerStarted struct {
	StartUp time.Duration
}

func (s *ServerStarted) Type() string {
	return "ServerStarted"
}

// [Server thread/INFO]: Stopping server
type ServerStopping struct {
}

func (s *ServerStopping) Type() string {
	return "ServerStopping"
}

// [Server thread/INFO]: **** joined the game
type PlayerJoined struct {
	PlayerName string
}

func (s *PlayerJoined) Type() string {
	return "PlayerJoined"
}

// [Server thread/INFO]: **** left the game
type PlayerLeft struct {
	PlayerName string
}

func (s *PlayerLeft) Type() string {
	return "PlayerLeft"
}

// [Async Chat Thread - #2/INFO]: <NAME> MESSAGE
type ChatMessage struct {
	PlayerName string
	Message    string
}

func (s *ChatMessage) Type() string {
	return "ChatMessage"
}

// [User Authenticator #26/INFO]: UUID of player kamijin is 98452fb7-def2-4253-8e4a-0ae57aa02604
type PlayerUUID struct {
	PlayerName string
	UUID       string
}

func (s *PlayerUUID) Type() string {
	return "PlayerUUID"
}
