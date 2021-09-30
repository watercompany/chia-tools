package telegrambot

import (
	"fmt"
	"net/http"
)

const (
	SendMessageFormat = "https://api.telegram.org/bot%s/sendMessage?chat_id=@%s&text=%s"
)

func SendMessage(botToken, chatID, msg string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf(SendMessageFormat, botToken, chatID, msg), nil)
	if err != nil {
		return fmt.Errorf("error in creating http request: %v", err)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error in executing http request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status code: %v", resp.StatusCode)
	}

	return nil
}
