package agent

import (
	"context"
	"encoding/json"
	"sync/atomic"
)

type Writer struct {
	Script string `json:"Script"`
	Total int32 `json:"Total"`
	Success int32 `json:"Success"`
	Failure int32 `json:"Failure"`
}

func NewWriter(script string) *Writer {
	return &Writer{
		Script: script,
	}
}

func (w *Writer) Start(ctx context.Context, ad *AgentData) {
	ad.Logger.Info("Writer starting")
	defer ad.ResultFile.Close()
	en := json.NewEncoder(ad.ResultFile)
	en.SetIndent("", "	")

	for {
		select {
		case <-ctx.Done():
			ad.Logger.Info("Writer shutting down on context done")
			if err := en.Encode(w); err != nil {
				ad.Logger.Errorf("Writer encode error on shutdown: %v", err)
			}
			return
		default:
		}
		
		select {
		case <-ctx.Done():
			ad.Logger.Info("Writer shutting down on context done")
			err := en.Encode(w)
			if err != nil {
				ad.Logger.Errorf("Writer encode error on shutdown: %v", err)
			}
			return
		case <-ad.success:
			atomic.AddInt32(&w.Success, 1)
			atomic.AddInt32(&w.Total, 1)
		case <-ad.fail:
			atomic.AddInt32(&w.Failure, 1)
			atomic.AddInt32(&w.Total, 1)
		}	
	}
}
