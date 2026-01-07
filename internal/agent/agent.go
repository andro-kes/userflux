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
	ad.wg.Wait()
}

// Запуск сценария со всеми настройками
func runScript(ctx context.Context, ad *AgentData) {
	defer ad.wg.Done()

	userId := ctx.Value("user_id")

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
		req, err := http.NewRequestWithContext(ctxFlow, method, url, nil)
		if err != nil {
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
			res["error"] = err
			ad.ch <- res
			return
		}
		cancel()
		
		dec := json.NewDecoder(resp.Body)
		resp.Body.Close()
		m := make(map[string]any) // result
		err = dec.Decode(&m)
		if err != nil {
			res["error"] = err
			ad.ch <- res
			return
		}
		res["result"] = m
		ad.ch <- res
	}
}
