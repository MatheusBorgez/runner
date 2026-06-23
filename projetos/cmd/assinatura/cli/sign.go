package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyriosdata/runner/internal/invoker"
	"github.com/kyriosdata/runner/internal/jdk"
	"github.com/kyriosdata/runner/internal/runtime"
	"github.com/spf13/cobra"
)

const defaultPort = 8080

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Cria uma assinatura digital simulada enviando os parâmetros ao assinador.jar.

Por padrão usa o modo servidor (HTTP). Use --local para invocação direta.

EXEMPLOS:
  assinatura sign --content $(echo -n "conteudo" | base64)
  assinatura sign --content <base64> --token meu-pin
  assinatura sign --content <base64> --local`,
	RunE: runSign,
}

func init() {
	signCmd.Flags().String("content", "", "Conteúdo a assinar em Base64 (obrigatório)")
	signCmd.Flags().String("token", "", "Token/PIN opcional para autenticação")
	signCmd.Flags().Int("port", defaultPort, "Porta do servidor assinador.jar")
	signCmd.Flags().Bool("local", false, "Usa invocação direta (subprocess) em vez do modo servidor")
	signCmd.MarkFlagRequired("content") //nolint:errcheck
	rootCmd.AddCommand(signCmd)
}

func runSign(cmd *cobra.Command, _ []string) error {
	content, _ := cmd.Flags().GetString("content")
	token, _ := cmd.Flags().GetString("token")
	port, _ := cmd.Flags().GetInt("port")
	local, _ := cmd.Flags().GetBool("local")

	inv, err := buildInvoker(port, local)
	if err != nil {
		return err
	}

	resp, err := inv.Sign(invoker.SignRequest{Content: content, Token: token})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(2)
	}

	printResponse(resp)
	if !resp.Valid {
		os.Exit(1)
	}
	return nil
}

func buildInvoker(port int, local bool) (*invoker.Invoker, error) {
	sm, err := runtime.NewStateManager()
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar estado: %w", err)
	}

	if !local {
		if state, err := sm.Load("assinador"); err == nil && runtime.PidAlive(state.PID) {
			baseURL := fmt.Sprintf("http://localhost:%d", state.Port)
			httpInv := invoker.NewHTTPInvoker(baseURL)
			if httpInv.Health() == nil {
				return httpInv, nil
			}
			sm.Remove("assinador") //nolint:errcheck
		}
	}

	// Modo local
	jdkProv := jdk.NewProvisioner(sm.JdkDir())
	javaPath, err := jdkProv.Ensure()
	if err != nil {
		return nil, fmt.Errorf("JDK não disponível: %w", err)
	}
	jarPath := sm.JarPath("assinador.jar")
	if _, statErr := os.Stat(jarPath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf(
			"assinador.jar não encontrado em %s\n"+
				"Dica: copie o assinador.jar para esse diretório ou use 'assinatura start'", jarPath)
	}
	return invoker.NewLocalInvoker(javaPath, jarPath), nil
}

func printResponse(resp *invoker.SignatureResponse) {
	out := struct {
		Signature string `json:"signature,omitempty"`
		Valid      bool   `json:"valid"`
		Message   string `json:"message"`
	}{resp.Signature, resp.Valid, resp.Message}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}
