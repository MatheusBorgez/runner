package cli

import (
	"fmt"
	"os"

	"github.com/kyriosdata/runner/internal/invoker"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Valida uma assinatura digital simulada enviando os parâmetros ao assinador.jar.

EXEMPLOS:
  assinatura validate --content <base64> --signature <assinatura>
  assinatura validate --content <base64> --signature <assinatura> --local
  assinatura validate --content <base64> --signature <assinatura> --port 9090`,
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().String("content", "", "Conteúdo original em Base64 (obrigatório)")
	validateCmd.Flags().String("signature", "", "Assinatura a validar (obrigatório)")
	validateCmd.Flags().Int("port", defaultPort, "Porta do servidor assinador.jar")
	validateCmd.Flags().Bool("local", false, "Usa invocação direta (subprocess) em vez do modo servidor")
	validateCmd.MarkFlagRequired("content")   //nolint:errcheck
	validateCmd.MarkFlagRequired("signature") //nolint:errcheck
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, _ []string) error {
	content, _ := cmd.Flags().GetString("content")
	signature, _ := cmd.Flags().GetString("signature")
	port, _ := cmd.Flags().GetInt("port")
	local, _ := cmd.Flags().GetBool("local")

	inv, err := buildInvoker(port, local)
	if err != nil {
		return err
	}

	resp, err := inv.Validate(invoker.ValidateRequest{Content: content, Signature: signature})
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
