package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestPayload struct {
	Ev     string `json:"ev"`
	Et     string `json:"et"`
	ID     string `json:"id"`
	UID    string `json:"uid"`
	MID    string `json:"mid"`
	T      string `json:"t"`
	P      string `json:"p"`
	L      string `json:"l"`
	SC     string `json:"sc"`
	ATRK1  string `json:"atrk1"`
	ATRV1  string `json:"atrv1"`
	ATRT1  string `json:"atrt1"`
	ATRK2  string `json:"atrk2"`
	ATRV2  string `json:"atrv2"`
	ATRT2  string `json:"atrt2"`
	ATRK3  string `json:"atrk3"`
	ATRV3  string `json:"atrv3"`
	ATRT3  string `json:"atrt3"`
	ATRK4  string `json:"atrk4"`
	ATRV4  string `json:"atrv4"`
	ATRT4  string `json:"atrt4"`
	UATRK1 string `json:"uatrk1"`
	UATRV1 string `json:"uatrv1"`
	UATRT1 string `json:"uatrt1"`
	UATRK2 string `json:"uatrk2"`
	UATRV2 string `json:"uatrv2"`
	UATRT2 string `json:"uatrt2"`
	UATRK3 string `json:"uatrk3"`
	UATRV3 string `json:"uatrv3"`
	UATRT3 string `json:"uatrt3"`
	UATRK4 string `json:"uatrk4"`
	UATRV4 string `json:"uatrv4"`
	UATRT4 string `json:"uatrt4"`
	UATRK5 string `json:"uatrk5"`
	UATRV5 string `json:"uatrv5"`
	UATRT5 string `json:"uatrt5"`
	UATRK6 string `json:"uatrk6"`
	UATRV6 string `json:"uatrv6"`
	UATRT6 string `json:"uatrt6"`
}

type TransformedPayload struct {
	Event           string               `json:"event"`
	EventType       string               `json:"event_type"`
	AppID           string               `json:"app_id"`
	UserID          string               `json:"user_id"`
	MessageID       string               `json:"message_id"`
	PageTitle       string               `json:"page_title"`
	PageURL         string               `json:"page_url"`
	BrowserLanguage string               `json:"browser_language"`
	ScreenSize      string               `json:"screen_size"`
	Attributes      map[string]Attribute `json:"attributes"`
	Traits          map[string]Trait     `json:"traits"`
}

type Attribute struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Trait struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

func transform_payload(originalPayload RequestPayload) TransformedPayload {
	return TransformedPayload{
		Event:           originalPayload.Ev,
		EventType:       originalPayload.Et,
		AppID:           originalPayload.ID,
		UserID:          originalPayload.UID,
		MessageID:       originalPayload.MID,
		PageTitle:       originalPayload.T,
		PageURL:         originalPayload.P,
		BrowserLanguage: originalPayload.L,
		ScreenSize:      originalPayload.SC,
		Attributes: map[string]Attribute{
			"button_text": {
				Value: originalPayload.ATRV1,
				Type:  originalPayload.ATRT1,
			},
			"color_variation": {
				Value: originalPayload.ATRV2,
				Type:  originalPayload.ATRT2,
			},
			"page_path": {
				Value: originalPayload.ATRV3,
				Type:  originalPayload.ATRT3,
			},
			"source": {
				Value: originalPayload.ATRV4,
				Type:  originalPayload.ATRT4,
			},
		},
		Traits: map[string]Trait{
			"user_score": {
				Value: originalPayload.UATRV1,
				Type:  originalPayload.UATRT1,
			},
			"gender": {
				Value: originalPayload.UATRV2,
				Type:  originalPayload.UATRT2,
			},
			"tracking_code": {
				Value: originalPayload.UATRV3,
				Type:  originalPayload.UATRT3,
			},
			"phone": {
				Value: originalPayload.UATRV4,
				Type:  originalPayload.UATRT4,
			},
			"coupon_clicked": {
				Value: originalPayload.UATRV5,
				Type:  originalPayload.UATRT5,
			},
			"opt_out": {
				Value: originalPayload.UATRV6,
				Type:  originalPayload.UATRT6,
			},
		},
	}
}

func ShareDatatoWebhook(payload TransformedPayload) error {
	webhookURL := "https://webhook.site/"

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Webhook request failed with status: %d", resp.StatusCode)
	}

	fmt.Println("Payload sent to webhook successfully")

	return nil
}

func worker(ch chan RequestPayload) {
	for {
		payload := <-ch
		fmt.Printf("Processing in worker:\n%+v\n", payload)

		transformedPayload := transform_payload(payload)
		err := ShareDatatoWebhook(transformedPayload)
		if err != nil {
			fmt.Println("Error sending to webhook:", err)
		}
	}
}

func handleRequest(ch chan RequestPayload, w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received request:\n%+v\n", payload)

	// Send the payload to the worker
	ch <- payload

	// Send a JSON response
	transformedPayload := transform_payload(payload)
	responseData, err := json.Marshal(transformedPayload)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func main() {
	requestChannel := make(chan RequestPayload)
	go worker(requestChannel)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(requestChannel, w, r)
	})

	port := 8080
	fmt.Printf("Server listening on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting the server:", err)
	}
}
