package main

import (
	"context"
	"strconv"
	"time"
	"net/http"
	"net/url"
	"log"
	"io"
	"fmt"
	"encoding/json"
)

type userData struct {
	Status string `json:"status"`
}

type usersGetResponse struct {
	Response []userData `json:"response"`
}

func requestStatus(client *http.Client, req *http.Request) (string, error) {
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("error requesting status: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}
	unm := &usersGetResponse{}
	err = json.Unmarshal(respBody, unm)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}
	return unm.Response[0].Status, nil
}

func launchPoller(ctx context.Context, accId int, vkToken string, pollingInterval time.Duration, newStatusChan chan<- string) error {
	client := &http.Client{}
	u := &url.URL{
		Scheme: "https",
		Host: "api.vk.com",
		Path: "method/users.get",
	}
	q := u.Query()
	q.Set("v", "5.199")
	q.Set("user_ids", strconv.Itoa(accId))
	q.Set("fields", "status")
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", vkToken))
	status, err := requestStatus(client, req)
	if err != nil {
		return err
	}
	log.Printf("successfully initialized poller with status: %s\n", status)
	go func() {
		ticker := time.NewTicker(pollingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				newStatus, err := requestStatus(client, req)
				if err != nil {
					log.Print(err)
					continue
				}
				if status != newStatus {
					status = newStatus
					newStatusChan<- status
				}

			}
		}
	}()
	return nil
}