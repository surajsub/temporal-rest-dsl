package executors

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/surajsub/temporal-rest-dsl/models"
	"io/ioutil"
	"net/http"
)

// HTTPExecutor implementation
type HTTPExecutor struct {
	BaseURL string
	Headers map[string]string
}

// NewHTTPExecutor creates an HTTPExecutor instance
func NewHTTPExecutor(config map[string]any) Executor {
	return &HTTPExecutor{
		BaseURL: config["baseURL"].(string),
		Headers: config["headers"].(map[string]string),
	}
}

func (h *HTTPExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {

	authToken, err := decrypt(h.Headers["Authorization"])
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt authorization token: %v", err)
	}
	h.Headers["Authorization"] = authToken.(string)

	// Validate the operation
	var mystep models.Step
	err = h.ValidateOperation(mystep)
	if err != nil {
		return nil, err.(error)
	}

	// Convert payload to JSON
	var requestBody []byte
	if payload != nil {
		requestBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	// Build the HTTP request
	url := h.BaseURL
	req, err := http.NewRequest(step.Operation, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}
	if step.Operation == "POST" || step.Operation == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response into a map
	var responseMap map[string]any
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return responseMap, nil
}

func decrypt(s string) (any, any) {
	key := []byte("thisisaverysecurekey32byteslong!") // Replace with your actual key
	ciphertext, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func (h *HTTPExecutor) ValidateOperation(step models.Step) error {
	supportedOperations := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}

	if !supportedOperations[step.Operation] {
		return fmt.Errorf("unsupported HTTP operation: %s", step.Operation)
	}
	return nil
}

// Register HTTPExecutor
func init() {
	RegisterExecutor("http", NewHTTPExecutor, []string{"GET", "POST", "PUT", "DELETE"})
}
