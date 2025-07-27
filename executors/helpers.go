package executors

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/providers"
	"golang.org/x/oauth2"
)

func Decrypt(cipherText, key string) (string, error) {
	// Ensure the key is 32 bytes (AES-256)
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("decryption key must be 32 bytes")
	}

	cipherData, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	if len(cipherData) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cipherData[:aes.BlockSize]
	cipherData = cipherData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherData, cipherData)

	return string(cipherData), nil
}

func CreateGitHubClient(token string) *github.Client {
	log.Printf("Creating GitHub client  with credentials")

	if token == "" {
		log.Fatalf("*************** TOKEN IS NOT AVAILABLE ********************")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// FormatVariables Utility function to format Terraform variable arguments
func FormatVariables(variables map[string]any) []string {
	var vars []string
	//var singleQuote = "'"
	for key, value := range variables {
		vars = append(vars, fmt.Sprintf("-var=%s=%v", key, value))
		//log.Printf("Formatted variables: %v", vars)
	}

	return vars
}

// FormatBicepVariables Utility function to format binary variables
func FormatBicepVariables(variables map[string]any) []string {
	var vars []string
	//var singleQuote = "'"
	for key, value := range variables {
		vars = append(vars, fmt.Sprintf("--parameters=%s=%v", key, value))

	}

	return vars
}

// Utility function to capture Terraform outputs
func CaptureTerraformOutputs(workspace string, logger *logrus.Logger) (map[string]any, error) {
	logger.Infof("Capturing Terraform outputs for workspace: %s", workspace)
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = workspace

	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture Terraform outputs: %w", err)
	}

	var rawOutputs map[string]map[string]any
	if err := json.Unmarshal(outputBytes, &rawOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform outputs: %w", err)
	}

	outputs := make(map[string]any)
	for key, details := range rawOutputs {
		if value, ok := details["value"]; ok {
			outputs[key] = value
		}
	}

	return outputs, nil
}

// Utility function to capture OpenTofu outputs
func CaptureOpenTofuOutputs(workspace string, logger *logrus.Logger) (map[string]any, error) {

	logger.Infof("Capturing Opentofu outputs for workspace: %s", workspace)
	cmd := exec.Command("tofu", "output", "-json")
	cmd.Dir = workspace

	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture OpenTofu outputs: %w", err)
	}

	var rawOutputs map[string]map[string]any
	if err := json.Unmarshal(outputBytes, &rawOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse OpenTofu outputs: %w", err)
	}

	outputs := make(map[string]any)
	for key, details := range rawOutputs {
		if value, ok := details["value"]; ok {
			outputs[key] = value
		}
	}

	return outputs, nil
}

// RunCommand

// Utility function to run shell commands
func OLDRunCommand(cmd *exec.Cmd, logger *logrus.Logger) error {

	logger.Infof("Executing command: %s in the directory %s", cmd.String(), cmd.Dir)
	out, err := cmd.CombinedOutput()
	logger.Debugf("Command output: %s", string(out))
	if err != nil {
		logger.Errorf("Command execution failed: %v", err)
		//return fmt.Errorf("command failed ", a ...any)

	}
	return err
}

func RunCommand(cmd *exec.Cmd, logger *logrus.Logger) error {
	var stderrBuf bytes.Buffer

	// Capture both stdout and stderr
	//cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	logger.Infof("Running command: %s", strings.Join(cmd.Args, " "))

	err := cmd.Run()

	//stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {

		logger.Errorf("Command failed: %v", err)
		if stderrStr != "" {
			stderrStr := strings.TrimSpace(stderrBuf.String())
			stderrStr = extractError(stderrStr)
			logger.Errorf("stderr: %s", stderrStr)
		}

		return fmt.Errorf("command failed: %w\nstderr: %s", err, stderrStr)
	}

	//if stdoutStr != "" {
	//	logger.Infof("stdout: %s", stdoutStr)
	//}

	return nil
}

// Get the Secrets

func GetSecretsProvider(providerType string, config map[string]string) (SecretsProvider, error) {
	switch providerType {
	case "vault":
		p := &providers.VaultSecretsProvider{}
		if err := p.Init(config); err != nil {
			return nil, err
		}

		return p, nil

	default:
		return nil, fmt.Errorf("unsupported secrets provider %s", providerType)
	}
}

func extractError(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Error:") {
			return strings.TrimSpace(line)
		}
	}
	// fallback: return trimmed last 10 lines as summary
	n := len(lines)
	if n >= 10 {
		return strings.TrimSpace(lines[n-10] + ": " + lines[n-1])
	}
	return strings.TrimSpace(stderr)
}
