package cli_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestVersionCommand valida que `go run . version` exibe "dev" (valor padrão sem ldflags).
func TestVersionCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	cmd.Dir = "../" // diretório cmd/assinatura/
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . version falhou: %v\nSaída: %s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "dev") {
		t.Errorf("esperado 'dev' na saída, obteve: %q", output)
	}
	if !strings.Contains(output, "assinatura") {
		t.Errorf("esperado 'assinatura' na saída, obteve: %q", output)
	}
}

// TestHelpCommand valida que --help retorna código 0 e exibe uso.
func TestHelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	cmd.Dir = "../"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run . --help falhou: %v\nSaída: %s", err, out)
	}
	if !strings.Contains(string(out), "assinatura") {
		t.Errorf("esperado 'assinatura' no help, obteve: %s", out)
	}
}
