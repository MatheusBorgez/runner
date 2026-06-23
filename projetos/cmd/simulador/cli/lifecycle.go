package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/kyriosdata/runner/internal/jdk"
	"github.com/kyriosdata/runner/internal/release"
	"github.com/kyriosdata/runner/internal/runtime"
	"github.com/spf13/cobra"
)

const (
	defaultSimuladorPort = 8443
	simuladorStateName   = "simulador"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o Simulador do HubSaúde",
	Long: `Inicia o simulador.jar em background.

O simulador.jar é baixado automaticamente se não estiver disponível localmente.
O JRE necessário também é provisionado automaticamente em ~/.hubsaude/.

EXEMPLOS:
  simulador start
  simulador start --source https://exemplo.com/simulador.jar`,
	RunE: runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o Simulador do HubSaúde",
	RunE:  runStop,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status do Simulador do HubSaúde",
	RunE:  runStatus,
}

func init() {
	startCmd.Flags().String("source", "", "URL alternativa para download do simulador.jar")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}

func runStart(cmd *cobra.Command, _ []string) error {
	source, _ := cmd.Flags().GetString("source")

	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	// Idempotência: reutiliza instância ativa
	if state, loadErr := sm.Load(simuladorStateName); loadErr == nil && runtime.PidAlive(state.PID) {
		if simuladorHealthy() {
			fmt.Printf("Simulador já está em execução na porta %d (PID %d).\n", state.Port, state.PID)
			return nil
		}
		sm.Remove(simuladorStateName) //nolint:errcheck
	}

	if portOccupied(defaultSimuladorPort) {
		return fmt.Errorf(
			"porta %d já está em uso por outro processo\n"+
				"Dica: verifique com 'simulador status' ou encerre o processo que ocupa a porta",
			defaultSimuladorPort)
	}

	jarPath, err := ensureSimuladorJar(sm, source)
	if err != nil {
		return err
	}

	javaPath, err := jdk.NewProvisioner(sm.JdkDir()).Ensure()
	if err != nil {
		return fmt.Errorf("JDK não disponível: %w", err)
	}

	proc := exec.Command(javaPath, "-jar", jarPath)
	if err := proc.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar simulador.jar: %w", err)
	}

	if err := sm.Save(simuladorStateName, runtime.ProcessState{
		PID: proc.Process.Pid, Port: defaultSimuladorPort, Name: simuladorStateName,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: não foi possível salvar estado: %v\n", err)
	}

	if err := waitSimuladorReady(30); err != nil {
		return fmt.Errorf("simulador iniciou mas não respondeu a tempo: %w\n"+
			"Dica: verifique os logs com 'simulador status'", err)
	}

	fmt.Printf("Simulador iniciado | porta %d | PID %d\n", defaultSimuladorPort, proc.Process.Pid)
	return nil
}

func runStop(_ *cobra.Command, _ []string) error {
	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	state, loadErr := sm.Load(simuladorStateName)
	if loadErr != nil && !simuladorHealthy() {
		fmt.Println("Simulador não está em execução.")
		return nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, shutErr := client.Post(
		fmt.Sprintf("http://localhost:%d/shutdown", defaultSimuladorPort),
		"application/json", nil,
	)
	if shutErr == nil {
		resp.Body.Close()
	}

	if loadErr == nil && runtime.PidAlive(state.PID) {
		if p, err := os.FindProcess(state.PID); err == nil {
			p.Kill() //nolint:errcheck
		}
	}

	sm.Remove(simuladorStateName) //nolint:errcheck
	fmt.Println("Simulador encerrado.")
	return nil
}

func runStatus(_ *cobra.Command, _ []string) error {
	sm, err := runtime.NewStateManager()
	if err != nil {
		return err
	}

	state, loadErr := sm.Load(simuladorStateName)
	if loadErr != nil || !runtime.PidAlive(state.PID) {
		if loadErr == nil {
			sm.Remove(simuladorStateName) //nolint:errcheck
		}
		fmt.Println("Simulador: não está em execução")
		return nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/api/info", defaultSimuladorPort))
	if err != nil {
		fmt.Printf("Simulador: processo ativo (PID %d) mas /api/info falhou: %v\n", state.PID, err)
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("Simulador: em execução | PID %d | porta %d | /api/info HTTP %d\n",
		state.PID, state.Port, resp.StatusCode)
	return nil
}

func simuladorHealthy() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/api/info", defaultSimuladorPort))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

func portOccupied(port int) bool {
	conn, err := http.Get(fmt.Sprintf("http://localhost:%d", port)) //nolint:gosec
	if err == nil {
		conn.Body.Close()
		return true
	}
	return false
}

func waitSimuladorReady(maxSeconds int) error {
	for i := 0; i < maxSeconds; i++ {
		if simuladorHealthy() {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("simulador não ficou pronto após %d segundos", maxSeconds)
}

func ensureSimuladorJar(sm *runtime.StateManager, source string) (string, error) {
	jarPath := sm.JarPath("simulador.jar")
	if source != "" {
		fmt.Fprintln(os.Stderr, "Baixando simulador.jar de fonte alternativa:", source)
		if err := release.DownloadFile(jarPath, source); err != nil {
			return "", fmt.Errorf("falha ao baixar simulador.jar: %w", err)
		}
		return jarPath, nil
	}
	return release.NewManager(sm.Dir()).EnsureSimulador()
}
