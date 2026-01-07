package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func Writer(ctx context.Context, ad *AgentData) {
	ad.Logger.Info("Writer starting")
	defer ad.ResultFile.Close()
	en := json.NewEncoder(ad.ResultFile)
	en.SetIndent("", "	")
	for {
		select {
		case res := <-ad.ch:
			err := en.Encode(res)
			if err != nil {
				ad.Logger.Errorf("Writer encode error: %v", err)
				fmt.Fprintln(os.Stderr, "writer encode error:", err)
			}
		case <-ctx.Done():
			ad.Logger.Info("Writer shutting down on context done")
			err := en.Encode("Agent is done")
			if err != nil {
				ad.Logger.Errorf("Writer encode error on shutdown: %v", err)
				fmt.Fprintln(os.Stderr, "writer encode error on shutdown:", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
