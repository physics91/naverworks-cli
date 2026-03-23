package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/cmd"
	"github.com/physics91/naverworks-cli/internal/api"
)

func main() {
	if err := cmd.Execute(); err != nil {
		errObj := map[string]map[string]string{
			"error": {"code": "CLI_ERROR", "description": err.Error()},
		}
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			errObj["error"]["code"] = apiErr.Code
			errObj["error"]["description"] = apiErr.Description
		}
		data, _ := json.Marshal(errObj)
		fmt.Fprintln(os.Stderr, string(data))
		os.Exit(1)
	}
}
