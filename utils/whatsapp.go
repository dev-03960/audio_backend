package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// SendUserCredentialWhatsApp sends user credentials via WhatsApp using Innosate API
func SendUserCredentialWhatsApp(name, phoneNumber, password string) error {
	apiKey := os.Getenv("INNOSATE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INNOSATE_API_KEY not set")
	}

	message := fmt.Sprintf("Hello %s! Your login credentials:\nPhone: %s\nPassword: %s", name, phoneNumber, password)

	payload := map[string]interface{}{
		"phoneNumber": phoneNumber,
		"message":     message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.innosate.com/send", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status: %d", resp.StatusCode)
	}

	return nil
}
