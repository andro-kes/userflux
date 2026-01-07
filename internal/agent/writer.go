package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func Writer(ctx context.Context, ad *AgentData) {
	en := json.NewEncoder(ad.ResultFile)
	en.SetIndent("", "	")
	for {
		select {
		case res := <-ad.ch:
			err := en.Encode(res)
			if err != nil {
				panic("Writer is unavailable")
			}
		case <-ctx.Done():
			err := en.Encode("Agent is done")
			if err != nil {
				panic("Writer is unavailable")
			}
			os.Exit(0)
		default:
			fmt.Println("Queue is full")
		}
	}
}
