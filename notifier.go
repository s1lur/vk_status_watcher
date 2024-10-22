package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"context"
	"net/url"
	"log"
	"time"
	backoff "github.com/cenkalti/backoff/v4"
)


type requestBody struct {
	ChatID int `json:"chat_id"`
	Text string `json:"text"`
}

func getSendMessageFunc(tgToken string, chatID int) func(text string) error {
	client := &http.Client{}
	u := &url.URL{
		Scheme: "https",
		Host: "api.telegram.org",
		Path: fmt.Sprintf("bot%s/sendMessage", tgToken),
	}

	return func (text string) error {
		reqBody, err := json.Marshal(requestBody{
			ChatID: chatID,
			Text: text,
		})
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", u.String(), bytes.NewReader(reqBody))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")

		return backoff.Retry(func() error {
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("error sending message: %v", err)
			}
			if resp.StatusCode == http.StatusOK {
				log.Printf("!!! message sent successfully !!!")
			} else {
				err := fmt.Errorf("received invalid status code: %d", resp.StatusCode)
				if resp.StatusCode < http.StatusInternalServerError {
					return backoff.Permanent(err)
				}
				return err
			} 
			return nil
		}, backoff.NewExponentialBackOff())
	}
}


func launchNotifier(ctx context.Context, chatID int, tgToken string, initStatus string, newStatusChan <-chan string) error {
	sendMessage := getSendMessageFunc(tgToken, chatID)
	curStatus := initStatus

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		statusChanged := false
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if h, m, _ := time.Now().Clock(); h == 9 && m == 0 {
					if !statusChanged {
						if err := sendMessage(fmt.Sprintf("Статус все ещё: %s", curStatus)); err != nil {
							log.Printf("error sending message: %v", err)
						}
					}
					statusChanged = false
				}
			case status := <-newStatusChan:
				statusChanged = true
				if err := sendMessage(fmt.Sprintf("Новый статус: %s", status)); err != nil {
					log.Printf("error sending message: %v", err)
				}
				curStatus = status
			}
		}
	}()
	return nil
}