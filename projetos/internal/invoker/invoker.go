package invoker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SignRequest struct {
	Content string `json:"content"`
	Token   string `json:"token,omitempty"`
}

type ValidateRequest struct {
	Content   string `json:"content"`
	Signature string `json:"signature"`
}

type SignatureResponse struct {
	Signature string `json:"signature"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
}

type Invoker struct {
	java    string
	jarPath string
	baseURL string
}

func NewLocalInvoker(javaPath, jarPath string) *Invoker {
	return &Invoker{java: javaPath, jarPath: jarPath}
}

func NewHTTPInvoker(baseURL string) *Invoker {
	return &Invoker{baseURL: strings.TrimRight(baseURL, "/")}
}

func (inv *Invoker) Sign(req SignRequest) (*SignatureResponse, error) {
	if inv.baseURL != "" {
		return inv.httpPost("/sign", req)
	}
	args := []string{"sign", "--content", req.Content}
	if req.Token != "" {
		args = append(args, "--token", req.Token)
	}
	return inv.runLocal(args)
}

func (inv *Invoker) Validate(req ValidateRequest) (*SignatureResponse, error) {
	if inv.baseURL != "" {
		return inv.httpPost("/validate", req)
	}
	return inv.runLocal([]string{"validate", "--content", req.Content, "--signature", req.Signature})
}

func (inv *Invoker) Health() error {
	if inv.baseURL == "" {
		return errors.New("Health() requer modo HTTP")
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(inv.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("servidor não acessível: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check retornou HTTP %d", resp.StatusCode)
	}
	return nil
}

func (inv *Invoker) Shutdown() error {
	if inv.baseURL == "" {
		return errors.New("Shutdown() requer modo HTTP")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(inv.baseURL+"/shutdown", "application/json", nil)
	if err != nil {
		return fmt.Errorf("erro ao solicitar shutdown: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (inv *Invoker) runLocal(args []string) (*SignatureResponse, error) {
	if _, err := os.Stat(inv.jarPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf(
			"assinador.jar não encontrado em %s\nDica: copie o assinador.jar para esse diretório", inv.jarPath)
	}

	cmd := exec.Command(inv.java, append([]string{"-jar", inv.jarPath}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return parseResponse(stdout.Bytes())
		}
		return nil, fmt.Errorf("erro ao executar assinador.jar (exit %d): %s", exitErr.ExitCode(), stderr.String())
	}
	return parseResponse(stdout.Bytes())
}

func (inv *Invoker) httpPost(path string, body any) (*SignatureResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(inv.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("erro na requisição HTTP: %w\nDica: verifique se o servidor está ativo com 'assinatura status'", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return parseResponse(respBody)
}

func parseResponse(data []byte) (*SignatureResponse, error) {
	var resp SignatureResponse
	if err := json.Unmarshal(bytes.TrimSpace(data), &resp); err != nil {
		return nil, fmt.Errorf("resposta inválida do assinador.jar: %w\nSaída: %s", err, string(data))
	}
	return &resp, nil
}
