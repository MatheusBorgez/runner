// Package jdk detecta e provisiona o JDK 21 (Temurin) em ~/.hubsaude/jdk/.
package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Provisioner detecta e baixa o JDK se necessário.
type Provisioner struct {
	jdkDir string
}

// NewProvisioner cria um Provisioner usando o diretório informado.
func NewProvisioner(jdkDir string) *Provisioner {
	return &Provisioner{jdkDir: jdkDir}
}

// Ensure garante que um JDK/JRE 21 está disponível e retorna o caminho do executável java.
func (p *Provisioner) Ensure() (string, error) {
	if java := p.localJava(); java != "" {
		return java, nil
	}
	if java, err := exec.LookPath("java"); err == nil {
		if ok, _ := isVersion21(java); ok {
			return java, nil
		}
	}
	fmt.Fprintln(os.Stderr, "JDK 21 não encontrado. Baixando Eclipse Temurin 21...")
	if err := p.download(); err != nil {
		return "", fmt.Errorf("falha ao baixar JDK: %w", err)
	}
	java := p.localJava()
	if java == "" {
		return "", fmt.Errorf("download concluído mas executável java não encontrado em %s", p.jdkDir)
	}
	return java, nil
}

func (p *Provisioner) localJava() string {
	for _, suffix := range []string{"bin/java", "bin/java.exe"} {
		if c := filepath.Join(p.jdkDir, suffix); fileExists(c) {
			return c
		}
	}
	entries, _ := os.ReadDir(p.jdkDir)
	for _, e := range entries {
		if e.IsDir() {
			for _, suffix := range []string{"bin/java", "bin/java.exe"} {
				if c := filepath.Join(p.jdkDir, e.Name(), suffix); fileExists(c) {
					return c
				}
			}
		}
	}
	return ""
}

func (p *Provisioner) download() error {
	url := downloadURL()
	if url == "" {
		return fmt.Errorf("plataforma não suportada: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if err := os.MkdirAll(p.jdkDir, 0o750); err != nil {
		return err
	}
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	archivePath := filepath.Join(p.jdkDir, "jdk21"+ext)
	fmt.Fprintf(os.Stderr, "Baixando de %s...\n", url)
	if err := downloadFile(archivePath, url); err != nil {
		return err
	}
	defer os.Remove(archivePath)
	fmt.Fprintln(os.Stderr, "Extraindo JDK...")
	if runtime.GOOS == "windows" {
		return extractZip(archivePath, p.jdkDir)
	}
	return extractTarGz(archivePath, p.jdkDir)
}

func downloadURL() string {
	archMap := map[string]string{"amd64": "x64", "arm64": "aarch64"}
	osMap := map[string]string{"linux": "linux", "darwin": "mac", "windows": "windows"}
	arch, okA := archMap[runtime.GOARCH]
	os_, okO := osMap[runtime.GOOS]
	if !okA || !okO {
		return ""
	}
	return fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/21/ga/%s/%s/jre/hotspot/normal/eclipse", os_, arch)
}

func isVersion21(javaPath string) (bool, error) {
	out, err := exec.Command(javaPath, "-version").CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(out), `version "21`), nil
}

func downloadFile(dest, url string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d ao baixar %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := safeJoin(destDir, header.Name)
		if target == "" {
			continue
		}
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0o750) //nolint:errcheck
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0o750) //nolint:errcheck
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target := safeJoin(destDir, f.Name)
		if target == "" {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0o750) //nolint:errcheck
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0o750) //nolint:errcheck
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func safeJoin(base, name string) string {
	target := filepath.Join(base, name)
	if !strings.HasPrefix(target, filepath.Clean(base)+string(os.PathSeparator)) {
		return ""
	}
	return target
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
