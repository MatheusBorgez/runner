package runtime_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kyriosdata/runner/internal/runtime"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	sm := stateManagerWithDir(t, dir)

	want := runtime.ProcessState{PID: 12345, Port: 8080, Name: "assinador"}
	if err := sm.Save("assinador", want); err != nil {
		t.Fatalf("Save falhou: %v", err)
	}

	got, err := sm.Load("assinador")
	if err != nil {
		t.Fatalf("Load falhou: %v", err)
	}
	if got != want {
		t.Errorf("esperado %+v, obteve %+v", want, got)
	}
}

func TestLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	sm := stateManagerWithDir(t, dir)

	_, err := sm.Load("inexistente")
	if err != runtime.ErrNotFound {
		t.Errorf("esperado ErrNotFound, obteve: %v", err)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	sm := stateManagerWithDir(t, dir)

	sm.Save("assinador", runtime.ProcessState{PID: 1, Port: 8080}) //nolint:errcheck
	sm.Remove("assinador")                                          //nolint:errcheck

	_, err := sm.Load("assinador")
	if err != runtime.ErrNotFound {
		t.Errorf("esperado ErrNotFound após Remove, obteve: %v", err)
	}
}

func TestRemoveIdempotent(t *testing.T) {
	dir := t.TempDir()
	sm := stateManagerWithDir(t, dir)
	if err := sm.Remove("naoexiste"); err != nil {
		t.Errorf("Remove de arquivo inexistente não deveria falhar: %v", err)
	}
}

func TestJarPath(t *testing.T) {
	dir := t.TempDir()
	sm := stateManagerWithDir(t, dir)
	expected := filepath.Join(dir, ".hubsaude", "assinador.jar")
	if got := sm.JarPath("assinador.jar"); got != expected {
		t.Errorf("JarPath: esperado %q, obteve %q", expected, got)
	}
}

// stateManagerWithDir cria um StateManager usando um diretório temporário.
func stateManagerWithDir(t *testing.T, dir string) *runtime.StateManager {
	t.Helper()
	// Cria manager apontando para dir injetando via hack de subdiretório
	// (StateManager usa os.UserHomeDir, então fazemos override via env HOME)
	t.Setenv("HOME", dir)
	sm, err := runtime.NewStateManager()
	if err != nil {
		t.Fatalf("NewStateManager falhou: %v", err)
	}
	_ = os.MkdirAll(filepath.Join(dir, ".hubsaude"), 0o750)
	return sm
}
