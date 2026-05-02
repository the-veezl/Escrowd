package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const paystackBase = "https://api.paystack.co"

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type VerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Status    string `json:"status"`
		Reference string `json:"reference"`
		Amount    int    `json:"amount"`
	} `json:"data"`
}

func InitializePayment(email string, amountKobo int, reference string, metadata map[string]string) (string, error) {
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("PAYSTACK_SECRET_KEY not set")
	}

	body := map[string]interface{}{
		"email":     email,
		"amount":    amountKobo,
		"reference": reference,
		"metadata":  metadata,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", paystackBase+"/transaction/initialize", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result InitializeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if !result.Status {
		return "", fmt.Errorf("paystack error: %s", result.Message)
	}

	return result.Data.AuthorizationURL, nil
}

func VerifyPayment(reference string) (bool, error) {
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if secretKey == "" {
		return false, errors.New("PAYSTACK_SECRET_KEY not set")
	}

	req, err := http.NewRequest("GET", paystackBase+"/transaction/verify/"+reference, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+secretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result VerifyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, err
	}

	return result.Data.Status == "success", nil
}
