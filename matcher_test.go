package main

import (
	"reflect"
	"testing"
	"time"
)

func TestParseLog(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    MessageVariant
		wantErr bool
	}{
		{
			name: "ServerStarting",
			args: "[Server thread/INFO]: Starting minecraft server version 1.19.2",
			want: &ServerStarting{
				Version: "1.19.2",
			},
			wantErr: false,
		},
		{
			name: "ServerStarted",
			args: "[Server thread/INFO]: Done (12.345s)! For help, type \"help\"",
			want: &ServerStarted{
				StartUp: time.Duration(12345 * time.Millisecond),
			},
			wantErr: false,
		},
		{
			name:    "ServerStopping",
			args:    "[Server thread/INFO]: Stopping server",
			want:    &ServerStopping{},
			wantErr: false,
		},
		{
			name: "PlayerJoined",
			args: "[Server thread/INFO]: player joined the game",
			want: &PlayerJoined{
				PlayerName: "player",
			},
			wantErr: false,
		},
		{
			name: "PlayerLeft",
			args: "[Server thread/INFO]: player left the game",
			want: &PlayerLeft{
				PlayerName: "player",
			},
		},
		{
			name: "ChatMessage",
			args: "[Async Chat Thread - #2/INFO]: <player> hello",
			want: &ChatMessage{
				PlayerName: "player",
				Message:    "hello",
			},
		},
		{
			name: "ChatMessage Japanese",
			args: "[Async Chat Thread - #0/INFO]: <kamijin> あああ",
			want: &ChatMessage{
				PlayerName: "kamijin",
				Message:    "あああ",
			},
		},
		{
			name: "UUID",
			args: "[User Authenticator #26/INFO]: UUID of player kamijin is 98452fb7-def2-4253-8e4a-0ae57aa02604",
			want: &PlayerUUID{
				PlayerName: "kamijin",
				UUID:       "98452fb7-def2-4253-8e4a-0ae57aa02604",
			},
		},
		{
			name:    "Unknown",
			args:    "[Server thread/INFO]: unknown",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLog(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
