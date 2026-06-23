// Package cli define os comandos do CLI simulador usando Cobra.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "CLI para gerenciamento do Simulador do HubSaúde",
	Long: `simulador — Sistema Runner | CLI do Simulador HubSaúde

Gerencia o ciclo de vida do simulador.jar (start, stop, status).
O JAR é baixado automaticamente na primeira execução e atualizado conforme release.json.

EXEMPLOS:
  simulador start
  simulador start --source https://exemplo.com/simulador.jar
  simulador stop
  simulador status
  simulador version`,
}

// Execute é o ponto de entrada do CLI simulador.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Exibe informações detalhadas de diagnóstico")
}
