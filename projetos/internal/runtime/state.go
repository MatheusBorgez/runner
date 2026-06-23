package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const hubsaudeDir = ".hubsaude"

type ProcessState struct {
	PID  int    `json:"pid"`
	Port int    `json:"port"`
	Name string `json:"name"`
}

type StateManager struct {
	dir string
}

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

func (m *StateManager) Save(name string, state ProcessState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar estado: %w", err)
	}
	return os.WriteFile(m.statePath(name), data, 0o640)
}

func (m *StateManager) Load(name string) (ProcessState, error) {
	data, err := os.ReadFile(m.statePath(name))
	if errors.Is(err, os.ErrNotExist) {
		return ProcessState{}, ErrNotFound
	}
	if err != nil {
		return ProcessState{}, fmt.Errorf("erro ao ler estado: %w", err)
	}
	var state ProcessState
	if err := json.Unmarshal(data, &state); err != nil {
		return ProcessState{}, fmt.Errorf("estado corrompido: %w", err)
	}
	return state, nil
}

func (m *StateManager) Remove(name string) error {
	err := os.Remove(m.statePath(name))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (m *StateManager) Dir() string    { return m.dir }
func (m *StateManager) JarPath(name string) string { return filepath.Join(m.dir, name) }
func (m *StateManager) JdkDir() string { return filepath.Join(m.dir, "jdk") }

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

var ErrNotFound = errors.New("estado não encontrado")

func (m *StateManager) statePath(name string) string {
	return filepath.Join(m.dir, name+".state.json")
}
