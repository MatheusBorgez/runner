// Package runtime gerencia o estado persistente dos processos em ~/.hubsaude/.
package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const hubsaudeDir = ".hubsaude"

// ProcessState representa o estado persistido de um processo gerenciado.
type ProcessState struct {
	PID  int    `json:"pid"`
	Port int    `json:"port"`
	Name string `json:"name"`
}

// StateManager gerencia arquivos de estado em ~/.hubsaude/.
type StateManager struct {
	dir string
}

// NewStateManager cria um StateManager usando ~/.hubsaude/ como diretório base.
func NewStateManager() (*StateManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("não foi possível determinar o diretório home: %w", err)
	}
	dir := filepath.Join(home, hubsaudeDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("não foi possível criar %s: %w", dir, err)
	}
	return &StateManager{dir: dir}, nil
}

// Save persiste o estado de um processo.
func (m *StateManager) Save(name string, state ProcessState) error {
	path := m.statePath(name)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar estado: %w", err)
	}
	return os.WriteFile(path, data, 0o640)
}

// Load lê o estado persistido de um processo. Retorna ErrNotFound se não existe.
func (m *StateManager) Load(name string) (ProcessState, error) {
	path := m.statePath(name)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ProcessState{}, ErrNotFound
	}
	if err != nil {
		return ProcessState{}, fmt.Errorf("erro ao ler estado: %w", err)
	}
	var state ProcessState
	if err := json.Unmarshal(data, &state); err != nil {
		return ProcessState{}, fmt.Errorf("estado corrompido em %s: %w", path, err)
	}
	return state, nil
}

// Remove apaga o arquivo de estado de um processo.
func (m *StateManager) Remove(name string) error {
	err := os.Remove(m.statePath(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Dir retorna o caminho base (~/.hubsaude/).
func (m *StateManager) Dir() string { return m.dir }

// JarPath retorna o caminho esperado de um JAR em ~/.hubsaude/.
func (m *StateManager) JarPath(name string) string { return filepath.Join(m.dir, name) }

// JdkDir retorna o diretório do JDK provisionado (~/.hubsaude/jdk/).
func (m *StateManager) JdkDir() string { return filepath.Join(m.dir, "jdk") }

// JavaExecutable retorna o executável java a ser usado: provisionado ou do PATH.
func (m *StateManager) JavaExecutable() string {
	jdkDir := m.JdkDir()
	for _, c := range []string{
		filepath.Join(jdkDir, "bin", "java"),
		filepath.Join(jdkDir, "bin", "java.exe"),
	} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "java"
}

// PidAlive verifica se um processo com o PID informado está vivo.
func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return isRunning(p)
}

// ErrNotFound indica que nenhum arquivo de estado foi encontrado.
var ErrNotFound = errors.New("estado não encontrado")

func (m *StateManager) statePath(name string) string {
	return filepath.Join(m.dir, name+".state.json")
}
