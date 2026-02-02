package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andro-kes/userflux/internal/session"
	"github.com/google/uuid"
)

// Структура для асинхронной работы агента
type AgentData struct {
	*session.Session
	start  time.Time
	ch     chan map[string]any
	wg     *sync.WaitGroup
	client *http.Client
}

type UserID string

func RunAgent(s *session.Session) {
	s.Logger.Infof("Agent starting... Time: %s", s.Time)
	ctx, cancel := context.WithTimeout(context.Background(), s.Time)
	defer cancel()

	client := http.Client{}

	size := int(s.Time.Seconds())

	ad := &AgentData{
		s,
		time.Now(),
		make(chan map[string]any, size),
		&sync.WaitGroup{},
		&client,
	}
	defer close(ad.ch)

	go Writer(ctx, ad)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	outer:
		for time.Since(ad.start) < s.Time {
			select {
			case <-ticker.C:
				ad.wg.Add(1)
				c := context.WithValue(ctx, "user_id", UserID(uuid.New().String()))
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
		res := map[string]any{
			"user_id":     userId,
			"script_name": ad.Data.Script.Name,
			"flow_name":   flow.Name,
			"result":      nil,
			"error":       nil,
		}
		ctxFlow, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		method := flow.Request.Method
		url := flow.URL + flow.Request.Path

		var jsonBody []byte
		if len(flow.Body) != 0 {
			body, err := generateBody(flow.Body)
			if err != nil {
				ad.Logger.Errorf("User %v request execution error: %v", userId, err)
				res["error"] = err
				ad.ch <- res
				return
			}
			jsonBody, err = json.Marshal(body)
			if err != nil {
				ad.Logger.Errorf("User %v request execution error: %v", userId, err)
				res["error"] = err
				ad.ch <- res
				return
			}
			ad.Logger.Infof("ID: %s, Username: %s, Password: %s", userId, body["username"], body["password"])
		}

		ad.Logger.Infof("User %v executing request: %s %s", userId, method, url)
		req, err := http.NewRequestWithContext(ctxFlow, method, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			ad.Logger.Errorf("User %v request creation error: %v", userId, err)
			res["error"] = err
			ad.ch <- res
			return
		}

		headers := flow.Request.Headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := ad.client.Do(req)
		if err != nil {
			ad.Logger.Errorf("User %v request execution error: %v", userId, err)
			res["error"] = err
			ad.ch <- res
			return
		}
		defer resp.Body.Close()
		cancel()

		dec := json.NewDecoder(resp.Body)
		m := make(map[string]any) // result
		err = dec.Decode(&m)
		if err != nil {
			ad.Logger.Errorf("User %v response decode error: %v", userId, err)
			res["error"] = err
			ad.ch <- res
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			ad.Logger.Errorf("User response status code error: %d", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			res["error"] = fmt.Sprintf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			ad.ch <- res
			return
		}
		resp.Body.Close()
		res["result"] = m
		ad.ch <- res
	}
	ad.Logger.Infof("User %v goroutine completed", userId)
}
