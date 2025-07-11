// cloudflareplus/validator.go
package cloudflareplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type validationRequest struct {
	Expression string `json:"expression"`
}

type validationResponse struct {
	Success bool      `json:"success"`
	Errors  []cfError `json:"errors"`
}

type cfError struct {
	Message string `json:"message"`
}

func ValidateExpression(expression string, accountID string, client *http.Client) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ruleset-expression/validate", accountID)

	payload, _ := json.Marshal(validationRequest{Expression: expression})
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result validationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("Cloudflare validation failed: %s", result.Errors[0].Message)
	}

	return nil
}
