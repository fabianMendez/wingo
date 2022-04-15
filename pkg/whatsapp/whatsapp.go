package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func SendMessage(to, subject, message string) error {
	waurl := os.Getenv("WA_URL")
	watoken := os.Getenv("WA_TOKEN")
	if waurl == "" {
		log.Print("Whatsapp URL not set")
		return nil
	}
	if watoken == "" {
		log.Print("Whatsapp token not set")
		return nil
	}

	request := struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}{
		To:      to,
		Message: fmt.Sprintf("*%s*\n\n%s", subject, message),
	}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(request)
	if err != nil {
		return fmt.Errorf("could not encode request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, waurl+"/send", body)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+watoken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
