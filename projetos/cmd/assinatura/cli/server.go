package cli

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kyriosdata/runner/internal/invoker"
	"github.com/kyriosdata/runner/internal/jdk"
	"github.com/kyriosdata/runner/internal/runtime"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o assinador.jar no modo servidor",
	Long: `Inicia o assinador.jar como servidor HTTP em background.

O PID e a porta são registrados em ~/.hubsaude/ para gestão posterior.

EXEMPLOS:
  assinatura start
  assinatura start --port 9090
  assinatura start --port 9090 --timeout 60`,
	RunE: runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o assinador.jar em execução",
	Long: `EXEMPLOS:
  assinatura stop
  assinatura stop --port 9090`,
	RunE: runStop,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status do assinador.jar",
	Long: `EXEMPLOS:
  assinatura status`,
	RunE: runStatus,
}

func init() {
	startCmd.Flags().Int("port", defaultPort, "Porta para o servidor assinador.jar")
	startCmd.Flags().Int("timeout", 0, "Minutos de inatividade antes do auto-shutdown (0 = desativado)")
	stopCmd.Flags().Int("port", defaultPort, "Porta do servidor a encerrar")
	statusCmd.Flags().Int("port", defaultPort, "Porta do servidor a verificar")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}

func runStart(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetInt("port")
	timeout, _ := cmd.Flags().GetInt("timeout")

	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	// Idempotência: reutiliza instância ativa
	if state, err := sm.Load("assinador"); err == nil && runtime.PidAlive(state.PID) {
		baseURL := fmt.Sprintf("http://localhost:%d", state.Port)
		if invoker.NewHTTPInvoker(baseURL).Health() == nil {
			fmt.Printf("assinador.jar já está em execução na porta %d (PID %d).\n", state.Port, state.PID)
			return nil
		}
		sm.Remove("assinador") //nolint:errcheck
	}

	jdkProv := jdk.NewProvisioner(sm.JdkDir())
	javaPath, err := jdkProv.Ensure()
	if err != nil {
		return fmt.Errorf("JDK não disponível: %w", err)
	}

	jarPath := sm.JarPath("assinador.jar")
	if _, statErr := os.Stat(jarPath); os.IsNotExist(statErr) {
		return fmt.Errorf("assinador.jar não encontrado em %s\nDica: copie o assinador.jar para esse diretório", jarPath)
	}

	args := []string{"-jar", jarPath, "server", "--port", fmt.Sprint(port)}
	if timeout > 0 {
		args = append(args, "--timeout", fmt.Sprint(timeout))
	}

	proc := exec.Command(javaPath, args...)
	if err := proc.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar assinador.jar: %w", err)
	}

	if err := sm.Save("assinador", runtime.ProcessState{PID: proc.Process.Pid, Port: port, Name: "assinador"}); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: não foi possível salvar estado: %v\n", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	if err := waitReady(invoker.NewHTTPInvoker(baseURL), 15); err != nil {
		return fmt.Errorf("servidor não ficou pronto: %w", err)
	}

	fmt.Printf("assinador.jar iniciado | porta %d | PID %d\n", port, proc.Process.Pid)
	return nil
}

func runStop(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetInt("port")

	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	shutErr := invoker.NewHTTPInvoker(baseURL).Shutdown()

	if shutErr != nil {
		state, loadErr := sm.Load("assinador")
		if loadErr != nil {
			fmt.Println("assinador.jar não está em execução na porta", port)
			return nil
		}
		p, _ := os.FindProcess(state.PID)
		if p != nil {
			p.Kill() //nolint:errcheck
		}
	}

	sm.Remove("assinador") //nolint:errcheck
	fmt.Printf("assinador.jar encerrado (porta %d).\n", port)
	return nil
}

func runStatus(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetInt("port")

	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	state, loadErr := sm.Load("assinador")
	if loadErr != nil {
		fmt.Println("assinador.jar: não está em execução")
		return nil
	}

	if !runtime.PidAlive(state.PID) {
		sm.Remove("assinador") //nolint:errcheck
		fmt.Printf("assinador.jar: processo registrado (PID %d) não está mais ativo\n", state.PID)
		return nil
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	if err := invoker.NewHTTPInvoker(baseURL).Health(); err != nil {
		fmt.Printf("assinador.jar: processo ativo (PID %d) mas health check falhou: %v\n", state.PID, err)
		return nil
	}

	fmt.Printf("assinador.jar: em execução | PID %d | porta %d | status OK\n", state.PID, state.Port)
	return nil
}

func waitReady(inv *invoker.Invoker, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		if inv.Health() == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("servidor não ficou pronto após %d tentativas", maxAttempts)
}
