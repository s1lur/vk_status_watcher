package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"context"
	"net/url"
	"log"
	backoff "github.com/cenkalti/backoff/v4"
)

type requestBody struct {
	ChatID int `json:"chat_id"`
	Text string `json:"text"`
}

func launchNotifier(ctx context.Context, chatID int, tgToken string, newStatusChan <-chan string) error {
	client := &http.Client{}
	u := &url.URL{
		Scheme: "https",
		Host: "api.telegram.org",
		Path: fmt.Sprintf("bot%s/sendMessage", tgToken),
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case status := <-newStatusChan:
				body, err := json.Marshal(requestBody{
					ChatID: chatID,
					Text: fmt.Sprintf("Новый статус: %s", status),
				})
				if err != nil {
					log.Printf("error marshalling request: %v", err)
					continue
				}

				req, err := http.NewRequest("GET", u.String(), bytes.NewReader(body))
				if err != nil {
					log.Printf("error creating request: %v", req)
					continue
				}
				req.Header.Add("Content-Type", "application/json")

				err = backoff.Retry(func() error {
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

				if err != nil {
					log.Print(err)
				}
			}
		}
	}()
	return nil
}