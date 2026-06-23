// Package cli define os comandos do CLI assinatura usando Cobra.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para assinatura digital via assinador.jar",
	Long: `assinatura — Sistema Runner | CLI de Assinatura Digital

Invoca o assinador.jar para criar e validar assinaturas digitais simuladas,
gerenciando automaticamente o JDK e o ciclo de vida do servidor.

MODOS DE OPERAÇÃO:
  Modo servidor (padrão): o assinador.jar fica em execução e o CLI usa HTTP.
  Modo local (--local):   cada invocação inicia o assinador.jar via subprocess.

EXEMPLOS:
  assinatura sign   --content <base64>
  assinatura validate --content <base64> --signature <assinatura>
  assinatura start  [--port 8080] [--timeout 30]
  assinatura stop   [--port 8080]
  assinatura status
  assinatura version`,
}

// Execute é o ponto de entrada do CLI assinatura.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Exibe informações detalhadas de diagnóstico")
}
