package telegrambot_test

import (
	"chia-tools/cli/Watcher/telegrambot"
	"testing"
)

func TestSendMessage(t *testing.T) {
	tests := []struct {
		scenario string
		token    string
		chatID   string
		message  string
	}{
		{
			scenario: "Send Normal Message",
			token:    "2014244690:AAFhWLw0fxwOhVHdIzdxmsVtNIYjqBEGRmA", // put bot token id here before starting test
			chatID:   "chiaupdateswatercompany",                        // put chat id here before starting test
			message:  "Test Message from tool.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			err := telegrambot.SendMessage(tc.token, tc.chatID, tc.message)
			if err != nil {
				t.Errorf("err in sending message:%v", err)
			}
		})
	}
}
