package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"zauto/internal/config"
	"zauto/internal/projectroot"
	"zauto/internal/store"
)

func main() {
	cfg := projectroot.ResolveConfig("config/config.json")
	wf, err := config.Load(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx := context.Background()
	st, err := store.OpenWorkflow(ctx, wf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer st.Close()

	params := config.DefaultPipelineParams([]string{"login", "auto_post", "fanpage_post"})
	params["fanpage_mode"] = "single"
	params["fanpage_index"] = float64(0)
	if err := st.SetAccountAutomation(ctx, 1, "facebook_pipeline", params, true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	acc, err := st.GetAccountByID(ctx, 1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, _ := json.MarshalIndent(map[string]interface{}{
		"id": acc.ID, "name": acc.Name, "flow": acc.AutomationFlow, "params": acc.AutomationParams, "fanpages": len(acc.Fanpages),
	}, "", "  ")
	fmt.Println(string(b))
}
