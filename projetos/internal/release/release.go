package release

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const ReleaseManifestURL = "https://raw.githubusercontent.com/kyriosdata/runner/main/release.json"

type Manifest struct {
	Jar       *ArtifactEntry `json:"jar,omitempty"`
	Simulador *ArtifactEntry `json:"simulador,omitempty"`
	Validador *ArtifactEntry `json:"validador,omitempty"`
	JRE       *JREEntry      `json:"jre,omitempty"`
}

type ArtifactEntry struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

type JREEntry struct {
	WindowsX64   string `json:"windows_x64"`
	WindowsArm64 string `json:"windows_arm64,omitempty"`
	LinuxX64     string `json:"linux_x64"`
	LinuxArm64   string `json:"linux_arm64,omitempty"`
	MacX64       string `json:"mac_x64"`
	MacArm64     string `json:"mac_arm64,omitempty"`
}

type Manager struct {
	manifestURL string
	cacheDir    string
}

func NewManager(cacheDir string) *Manager {
	return &Manager{manifestURL: ReleaseManifestURL, cacheDir: cacheDir}
}

func (m *Manager) FetchManifest() (*Manifest, error) {
	resp, err := http.Get(m.manifestURL) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar release.json: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d ao buscar release.json", resp.StatusCode)
	}
	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("erro ao parsear release.json: %w", err)
	}
	return &manifest, nil
}

func (m *Manager) JREDownloadURL(manifest *Manifest) (string, error) {
	if manifest.JRE == nil {
		return "", fmt.Errorf("release.json não contém entradas de JRE")
	}
	goos, goarch := runtime.GOOS, runtime.GOARCH
	switch {
	case goos == "linux" && goarch == "amd64":
		return manifest.JRE.LinuxX64, nil
	case goos == "linux" && goarch == "arm64":
		if manifest.JRE.LinuxArm64 == "" {
			return "", fmt.Errorf("JRE não disponível para linux/arm64 no release.json")
		}
		return manifest.JRE.LinuxArm64, nil
	case goos == "darwin" && goarch == "arm64":
		if manifest.JRE.MacArm64 != "" {
			return manifest.JRE.MacArm64, nil
		}
		return manifest.JRE.MacX64, nil
	case goos == "darwin":
		return manifest.JRE.MacX64, nil
	case goos == "windows" && goarch == "arm64":
		if manifest.JRE.WindowsArm64 != "" {
			return manifest.JRE.WindowsArm64, nil
		}
		return manifest.JRE.WindowsX64, nil
	case goos == "windows":
		return manifest.JRE.WindowsX64, nil
	default:
		return "", fmt.Errorf("plataforma não suportada: %s/%s", goos, goarch)
	}
}

// EnsureSimulador baixa o simulador.jar apenas se a versão local for diferente da remota.
func (m *Manager) EnsureSimulador() (string, error) {
	manifest, err := m.FetchManifest()
	if err != nil {
		return "", err
	}
	if manifest.Simulador == nil {
		return "", fmt.Errorf("release.json não contém entrada 'simulador'")
	}

	jarPath := filepath.Join(m.cacheDir, "simulador.jar")
	versionPath := filepath.Join(m.cacheDir, "simulador.version")

	if localVersion, err := os.ReadFile(versionPath); err == nil {
		if strings.TrimSpace(string(localVersion)) == manifest.Simulador.Version {
			if _, err := os.Stat(jarPath); err == nil {
				fmt.Fprintf(os.Stderr, "simulador.jar v%s já disponível localmente.\n", manifest.Simulador.Version)
				return jarPath, nil
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Baixando simulador.jar v%s...\n", manifest.Simulador.Version)
	if err := os.MkdirAll(m.cacheDir, 0o750); err != nil {
		return "", err
	}
	if err := DownloadFile(jarPath, manifest.Simulador.URL); err != nil {
		return "", fmt.Errorf("erro ao baixar simulador.jar: %w", err)
	}
	os.WriteFile(versionPath, []byte(manifest.Simulador.Version), 0o640) //nolint:errcheck
	return jarPath, nil
}

func DownloadFile(dest, url string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
