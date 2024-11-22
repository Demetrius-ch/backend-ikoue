// backend/main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

type EmailRequest struct {
	Nom     string `json:"nom"`
	Prenom  string `json:"prenom"`
	Adresse string `json:"adresse"`
	Message string `json:"message"`
}

func sendSMTP(emailRequest EmailRequest) error {
	auth := smtp.PlainAuth("", "user@example.com", "password", "localhost")
	to := []string{"destination@example.com"}
	msg := []byte("Subject: Message de " + emailRequest.Nom + "\r\n\r\n" + emailRequest.Message)
	err := smtp.SendMail("localhost:1025", auth, "sender@example.com", to, msg)
	return err
}

func sendSendGrid(emailRequest EmailRequest) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	requestBody := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": "destination@example.com"}}},
		},
		"from":    map[string]string{"email": "sender@example.com"},
		"subject": "Message de " + emailRequest.Nom,
		"content": []map[string]string{{"type": "text/plain", "value": emailRequest.Message}},
	}

	requestBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	var emailRequest EmailRequest
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &emailRequest)

	err := sendSMTP(emailRequest)
	if err != nil {
		log.Println("SMTP error:", err)
		sendSendGrid(emailRequest)
	}

	response := map[string]string{"message": "Email envoyé avec succès."}
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/send-email", handler)
	fmt.Println("Back-end running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
