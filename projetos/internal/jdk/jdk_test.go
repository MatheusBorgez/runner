package jdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsurePrefersLocalJava(t *testing.T) {
	dir := t.TempDir()
	java := filepath.Join(dir, "bin", "java")
	if err := os.MkdirAll(filepath.Dir(java), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(java, []byte("não executa"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := NewProvisioner(dir).Ensure()
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if got != java {
		t.Fatalf("esperado %s, obteve %s", java, got)
	}
}

func TestJava21OnPath(t *testing.T) {
	dir := t.TempDir()
	java := filepath.Join(dir, "java")
	script := "#!/bin/sh\nprintf '%s\\n' 'openjdk version \"21.0.1\"' >&2\n"
	if err := os.WriteFile(java, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	if got := java21OnPath(); got != java {
		t.Fatalf("esperado %s, obteve %q", java, got)
	}
}

func TestJava21OnPathRejectsOtherVersion(t *testing.T) {
	dir := t.TempDir()
	java := filepath.Join(dir, "java")
	script := "#!/bin/sh\nprintf '%s\\n' 'openjdk version \"17.0.1\"' >&2\n"
	if err := os.WriteFile(java, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	if got := java21OnPath(); got != "" {
		t.Fatalf("versão 17 não deveria ser aceita, obteve %s", got)
	}
}

func TestJava21OnPathMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if got := java21OnPath(); got != "" {
		t.Fatalf("PATH sem java deveria retornar vazio, obteve %s", got)
	}
}
