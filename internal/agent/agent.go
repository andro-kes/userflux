package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/andro-kes/userflux/internal/session"
)

type AgentData struct {
	*session.Session
	ch     chan map[string]any
	wg     *sync.WaitGroup
	client *http.Client
}

func RunAgent(s *session.Session) {
	s.Logger.Infof("Agent starting with %d users for %s", s.Users, s.Time)
	ctx, cancel := context.WithTimeout(context.Background(), s.Time)
	defer cancel()

	client := http.Client{}

	ad := &AgentData{
		s,
		make(chan map[string]any, s.Users*len(s.Data.Script.Flow)),
		&sync.WaitGroup{},
		&client,
	}

	go Writer(ctx, ad)

	for i := 0; i < s.Users; i++ {
		ad.wg.Add(1)
		ctx := context.WithValue(ctx, "user_id", i)
		go runScript(ctx, ad)
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
		ctxFlow, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		method := flow.Request.Method
		url := flow.URL + flow.Request.Path
		ad.Logger.Infof("User %v executing request: %s %s", userId, method, url)
		req, err := http.NewRequestWithContext(ctxFlow, method, url, nil)
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
		resp.Body.Close()
		res["result"] = m
		ad.ch <- res
	}
	ad.Logger.Infof("User %v goroutine completed", userId)
}
