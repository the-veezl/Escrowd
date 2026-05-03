package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	sandboxBase = "https://sandbox.safaricom.co.ke"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

type stkPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type QueryResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResultCode          string `json:"ResultCode"`
	ResultDesc          string `json:"ResultDesc"`
}

func getAccessToken() (string, error) {
	consumerKey := os.Getenv("MPESA_CONSUMER_KEY")
	consumerSecret := os.Getenv("MPESA_CONSUMER_SECRET")

	if consumerKey == "" || consumerSecret == "" {
		return "", errors.New("MPESA_CONSUMER_KEY or MPESA_CONSUMER_SECRET not set")
	}

	auth := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))

	req, err := http.NewRequest("GET", sandboxBase+"/oauth/v1/generate?grant_type=client_credentials", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("empty access token received")
	}

	return tokenResp.AccessToken, nil
}

func generatePassword(shortCode string, passkey string, timestamp string) string {
	raw := shortCode + passkey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func InitiateSTKPush(phone string, amount int, escrowID string) (*STKPushResponse, error) {
	shortCode := os.Getenv("MPESA_SHORTCODE")
	passkey := os.Getenv("MPESA_PASSKEY")

	if shortCode == "" || passkey == "" {
		return nil, errors.New("MPESA_SHORTCODE or MPESA_PASSKEY not set")
	}

	token, err := getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	password := generatePassword(shortCode, passkey, timestamp)

	payload := stkPushRequest{
		BusinessShortCode: shortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            amount,
		PartyA:            phone,
		PartyB:            shortCode,
		PhoneNumber:       phone,
		CallBackURL:       "https://escrowd.fly.dev/mpesa/callback",
		AccountReference:  "escrowd-" + escrowID[:8],
		TransactionDesc:   "Escrow deal payment",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", sandboxBase+"/mpesa/stkpush/v1/processrequest", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("STK push failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, err
	}

	if stkResp.ResponseCode != "0" {
		return nil, fmt.Errorf("STK push error: %s", stkResp.ResponseDescription)
	}

	return &stkResp, nil
}

func QuerySTKStatus(checkoutRequestID string) (*QueryResponse, error) {
	shortCode := os.Getenv("MPESA_SHORTCODE")
	passkey := os.Getenv("MPESA_PASSKEY")

	token, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("20060102150405")
	password := generatePassword(shortCode, passkey, timestamp)

	payload := map[string]string{
		"BusinessShortCode": shortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"CheckoutRequestID": checkoutRequestID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", sandboxBase+"/mpesa/stkpushquery/v1/query", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, err
	}

	return &queryResp, nil
}

func FormatPhone(phone string) string {
	if len(phone) == 10 && phone[0] == '0' {
		return "254" + phone[1:]
	}
	if len(phone) == 12 && phone[:3] == "254" {
		return phone
	}
	if len(phone) == 9 {
		return "254" + phone
	}
	return phone
}
