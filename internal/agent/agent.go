package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/andro-kes/userflux/internal/session"
	"github.com/google/uuid"
)

// Структура для асинхронной работы агента
type AgentData struct {
	*session.Session
	start time.Time
	success     chan struct{}
	fail chan struct{}
	wg     *sync.WaitGroup
	client *http.Client
}

func RunAgent(s *session.Session) {
	s.Logger.Infof("Agent starting... Time: %s", s.Time)
	ctx, cancel := context.WithTimeout(context.Background(), s.Time)
	defer cancel()

	client := http.Client{}

	size := 100

	ad := &AgentData{
		s,
		time.Now(),
		make(chan struct{}, size),
		make(chan struct{}, size),
		&sync.WaitGroup{},
		&client,
	}

	writer := NewWriter(s.Data.Script.Name)
	go writer.Start(ctx, ad)

	outer:
	for time.Since(ad.start) < s.Time {
		delay := RandomDelay(time.Second, "uniform", 0.3, 50*time.Millisecond, 10*time.Second)
		timer := time.NewTimer(delay)

		select {
		case <-timer.C:
			ad.wg.Add(1)
			c := context.WithValue(ctx, "user_id", uuid.New().String())
			go runScript(c, ad)
		case <-ctx.Done():
			ad.Logger.Info("Agent was expired")
			break outer
		}
	}

	s.Logger.Info("Waiting for all user goroutines to complete")
	ad.wg.Wait()
	s.Logger.Info("All user goroutines completed")
}

// Запуск сценария со всеми настройками
func runScript(ctx context.Context, ad *AgentData) {
	defer ad.wg.Done()

	userId := ctx.Value("user_id")
	ad.Logger.Infof("User %v goroutine starting", userId)

	for _, flow := range ad.Data.Script.Flow {
		method := flow.Request.Method
		url := flow.URL + flow.Request.Path

		var jsonBody []byte
		if len(flow.Body) != 0 {
			body, err := generateBody(flow.Body)
			if err != nil {
				ad.Logger.Errorf("User %v request execution error: %v", userId, err)
				ad.fail <- struct{}{}
				return
			}
			jsonBody, err = json.Marshal(body)
			if err != nil {
				ad.Logger.Errorf("User %v request execution error: %v", userId, err)
				ad.fail <- struct{}{}
				return
			}
			ad.Logger.Infof("ID: %s, Username: %s, Password: %s", userId, body["username"], body["password"])
		}
		
		ad.Logger.Infof("User %v executing request: %s %s", userId, method, url)
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			ad.Logger.Errorf("User %v request creation error: %v", userId, err)
			ad.fail <- struct{}{}
			return
		}

		headers := flow.Request.Headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := ad.client.Do(req)
		if err != nil {
			ad.Logger.Errorf("User %v request execution error: %v", userId, err)
			ad.fail <- struct{}{}
			return
		}
		defer resp.Body.Close()
		
		dec := json.NewDecoder(resp.Body)
		m := make(map[string]any) // result
		err = dec.Decode(&m)
		if err != nil {
			ad.Logger.Errorf("User %v response decode error: %v", userId, err)
			ad.fail <- struct{}{}
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			ad.Logger.Errorf("User response status code error: %d", resp.StatusCode)
			ad.fail <- struct{}{}
			return
		}
		resp.Body.Close()
		ad.success <- struct{}{}
	}
	ad.Logger.Infof("User %v goroutine completed", userId)
}
