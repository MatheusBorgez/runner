# Refatorações sugeridas

Seis refatorações para os PRs da etapa 5. Nenhuma delas está aplicada nesta branch. Cada uma cabe num PR pequeno, feito localmente, com uma técnica do catálogo de Martin Fowler (*Refactoring*, 2ª edição, quando o nome existe lá; a 3 usa o nome clássico da 1ª edição).

Base de cada branch: `main` **depois** que o PR da pipeline estiver mergeado. Assim o `build.yml` (testes, golangci-lint e smoke) roda no PR. Enquanto esse PR não entrar, dá para ramificar a partir da branch da pipeline.

Não apontar os PRs para o upstream `kyriosdata/runner`. Base: `main` deste fork.

## Ordem e paralelismo

As seis mexem em arquivos diferentes. Duas pessoas podem fazê-las ao mesmo tempo, desde que cada uma fique com um número.

| Onda | Pessoa A | Pessoa B | Por que não colidem |
|---|---|---|---|
| 1 | 1. Remove Dead Code | 2. Guard clauses | `projetos/assinatura/` e `internal/invoker/` |
| 2 | 3. Magic number (Java) | 4. Extract Function (`release.go`) | Java e `internal/release/` |
| 3 | 5. Decompose Conditional (`jdk.go`) | 6. Extract Function (espera) | `internal/jdk/jdk.go` e `server.go` + `lifecycle.go` + arquivo novo |

Se só uma pessoa for fazendo, a ordem 1 → 2 → 3 → 4 → 5 → 6 é a mais segura: as primeiras são as menores.

Não juntar duas técnicas no mesmo PR. Não corrigir, no PR 5, os `errcheck` de `jdk.go`: isso mistura o diff com o PR 4 e aumenta a chance de conflito se alguém também abrir `jdk.go`.

O workflow legado `.github/workflows/ci.yml` sai na branch da pipeline. O PR 1 não precisa apagá-lo de novo. Se a branch da refatoração ainda tiver o arquivo, apague-o nesse PR e diga isso na descrição.

## Como testar (vale para todas)

```bash
cd projetos
go test ./... -count=1
go vet ./...
golangci-lint run ./... --timeout 8m
```

Versão usada na baseline: **golangci-lint 2.15.0**. Sem arquivo de configuração. O comando acima falha com código 1 enquanto existirem as 18 issues `errcheck` da baseline. Isso é esperado. Confira que o PR não aumentou o total (o relatório no fim da saída) e que `go test` passou. O CI desta disciplina não quebra o build por essas issues antigas.

Java, só no PR 3:

```bash
cd projetos/assinador-java
mvn -B test
```

---

## 1. Remove Dead Code

**Técnica:** Remove Dead Code (Fowler, *Refactoring*, 2ª ed.).

**Problema / smell:** código morto. `projetos/assinatura/` é um módulo antigo (`github.com/kyriosdata/assinatura`) com `main.go`, `cmd/root.go` e `cmd/version.go`. A CLI atual é `projetos/cmd/assinatura`. Nada no módulo `github.com/kyriosdata/runner` importa o módulo antigo. O `ci.yml` legado compilava esse diretório.

**Arquivos:** o diretório inteiro:

- `projetos/assinatura/main.go`
- `projetos/assinatura/cmd/root.go`
- `projetos/assinatura/cmd/version.go`
- `projetos/assinatura/go.mod`
- `projetos/assinatura/go.sum`
- `projetos/assinatura/README.md`

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/remove-dead-code-assinatura
git rm -r projetos/assinatura
```

Procure menções restantes (não apague documentação histórica da disciplina se a frase só cita o nome do comando `assinatura`):

```bash
git grep -n "projetos/assinatura" || true
```

Se `git grep` achar referência operacional a esse diretório, atualize a frase para `projetos/cmd/assinatura`. Rode `go test ./...` em `projetos/` (o módulo vivo não depende do diretório removido).

**Título sugerido:** `refactor(remove dead code): remove o módulo legado projetos/assinatura`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Remove Dead Code (Fowler, Refactoring, 2ª ed.)

## Motivação
`projetos/assinatura` é o módulo `github.com/kyriosdata/assinatura`, substituído por `projetos/cmd/assinatura`. Nenhum import do módulo atual o utiliza.

## O que mudou
- Removido o diretório `projetos/assinatura/`.

## Como testar
`cd projetos && go test ./... -count=1`
```

**Conflito:** nenhum com as outras cinco. Não edite arquivos Go do módulo atual.

---

## 2. Replace Nested Conditional with Guard Clauses

**Técnica:** Replace Nested Conditional with Guard Clauses (Fowler, *Refactoring*, 2ª ed.).

**Problema / smell:** condicional frágil; em caso de erro pode haver panic. Em `runLocal`, se `cmd.Run()` falha com um erro que não é `*exec.ExitError`, `errors.As` não preenche `exitErr`. A linha seguinte chama `exitErr.ExitCode()` com `exitErr == nil`.

Trecho atual (`projetos/internal/invoker/invoker.go`, linhas 104–110):

```go
err := cmd.Run()
if err != nil {
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
        return parseResponse(stdout.Bytes())
    }
    return nil, fmt.Errorf("erro ao executar assinador.jar (exit %d): %s", exitErr.ExitCode(), stderr.String())
}
```

O golangci-lint da baseline não aponta esse trecho. A issue `errcheck` desse arquivo está na linha 73 (`resp.Body.Close`) e **não** entra neste PR.

**Arquivos:**

- `projetos/internal/invoker/invoker.go` (só `runLocal`)
- `projetos/internal/invoker/invoker_test.go` (teste novo no fim)

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/guard-clauses-invoker
```

Substitua o bloco acima por cláusulas de guarda:

```go
err := cmd.Run()
if err != nil {
    var exitErr *exec.ExitError
    if !errors.As(err, &exitErr) {
        return nil, fmt.Errorf("erro ao executar assinador.jar: %w", err)
    }
    if exitErr.ExitCode() == 1 {
        return parseResponse(stdout.Bytes())
    }
    return nil, fmt.Errorf("erro ao executar assinador.jar (exit %d): %s", exitErr.ExitCode(), stderr.String())
}
```

O comportamento de exit code 1 (ainda parseia o stdout) e dos demais exit codes permanece. Só o caminho que não é `ExitError` deixa de entrar em panic.

Acrescente em `invoker_test.go` os imports `os` e `path/filepath`, e este teste:

```go
func TestRunLocalErroQueNaoEExitNaoEntraEmPanic(t *testing.T) {
    dir := t.TempDir()
    jar := filepath.Join(dir, "assinador.jar")
    if err := os.WriteFile(jar, []byte("nao-e-zip"), 0o644); err != nil {
        t.Fatal(err)
    }

    inv := invoker.NewLocalInvoker(filepath.Join(dir, "java-que-nao-existe"), jar)
    defer func() {
        if rec := recover(); rec != nil {
            t.Fatalf("panic inesperado: %v", rec)
        }
    }()

    _, err := inv.Sign(invoker.SignRequest{Content: "YQ=="})
    if err == nil {
        t.Fatal("esperado erro ao executar binário inexistente")
    }
}
```

Antes da correção esse teste entra em panic. Depois, `go test ./internal/invoker/ -count=1 -run TestRunLocalErroQueNaoEExitNaoEntraEmPanic` passa.

**Título sugerido:** `refactor(guard clauses): evita panic quando o JAR não termina com ExitError`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Replace Nested Conditional with Guard Clauses (Fowler, Refactoring, 2ª ed.)

## Motivação
Em `runLocal`, um erro que não é `*exec.ExitError` deixa `exitErr` nulo e a chamada a `ExitCode()` entra em panic.

## O que mudou
- `projetos/internal/invoker/invoker.go`: guarda antes de usar `exitErr`.
- Teste em `invoker_test.go` com binário inexistente (erro `*exec.Error`, não `*exec.ExitError`).

## Como testar
`cd projetos && go test ./internal/invoker/ -count=1`
```

**Conflito:** baixo com as outras. Quem for mexer na linha 73 (`Body.Close`) do mesmo arquivo deve esperar este PR ou deixar essa linha quieta.

---

## 3. Replace Magic Number with Symbolic Constant

**Técnica:** Replace Magic Number with Symbolic Constant (Fowler, *Refactoring*, 1ª ed.; na 2ª ed. o efeito equivalente é extrair o literal para um nome).

**Problema / smell:** números mágicos. `SignatureServer` repete 200, 405 e 422 nos handlers HTTP.

**Arquivos e linhas:** `projetos/assinador-java/src/main/java/com/kyriosdata/assinador/http/SignatureServer.java`

- linha 105: `405`
- linha 110: `200` e `422`
- linha 115: `405`
- linha 120: `200` e `422`
- linha 126: `200` (`/health`)
- linha 131: `405`
- linha 134: `200` (`/shutdown`)

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/magic-number-signature-server
```

No topo da classe, junto dos outros `static final`, acrescente:

```java
private static final int HTTP_OK = 200;
private static final int HTTP_METHOD_NOT_ALLOWED = 405;
private static final int HTTP_UNPROCESSABLE_ENTITY = 422;
```

Troque só esses literais nos quatro handlers. Não mude o JSON, a porta nem o fluxo. Rode `mvn -B test` em `projetos/assinador-java`. Os testes de `/health`, `/sign` e `/validate` já cobrem esses status.

**Título sugerido:** `refactor(magic number): nomeia os status HTTP do SignatureServer`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Replace Magic Number with Symbolic Constant (Fowler, Refactoring)

## Motivação
Os handlers de `SignatureServer` repetem 200, 405 e 422 sem nome.

## O que mudou
- Constantes `HTTP_OK`, `HTTP_METHOD_NOT_ALLOWED` e `HTTP_UNPROCESSABLE_ENTITY`.
- Nenhuma mudança de comportamento.

## Como testar
`cd projetos/assinador-java && mvn -B test`
```

**Conflito:** nenhum com as refatorações Go. Não implemente PKCS#11 de verdade neste PR.

---

## 4. Extract Function

**Técnica:** Extract Function (Fowler, *Refactoring*, 2ª ed.).

**Problema / smell:** o mesmo fechamento de recurso, com erro ignorado, repetido em `release.go`. A baseline marca quatro `errcheck` neste arquivo (linhas 53, 134, 142 e 152). Há ainda um `WriteFile` ignorado na linha 125 (`//nolint:errcheck`).

**Arquivo:** `projetos/internal/release/release.go` apenas. Não abra `jdk.go` aqui (são oito `errcheck` parecidos; outro PR, depois, para não colidir com a refatoração 5).

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/extract-function-release-close
```

Extraia uma função no mesmo pacote:

```go
func closeChecked(c io.Closer) error {
    if err := c.Close(); err != nil {
        return fmt.Errorf("erro ao fechar recurso: %w", err)
    }
    return nil
}
```

`FetchManifest` e `DownloadFile` precisam de retorno nomeado para o `defer` poder registrar o erro do `Close` sem esconder o erro principal. Exemplo para `DownloadFile`:

```go
func DownloadFile(dest, url string) (err error) {
    resp, err := http.Get(url) //nolint:gosec
    if err != nil {
        return fmt.Errorf("erro na requisição: %w", err)
    }
    defer func() {
        if cerr := closeChecked(resp.Body); cerr != nil && err == nil {
            err = cerr
        }
    }()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    f, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer func() {
        if cerr := closeChecked(f); cerr != nil && err == nil {
            err = cerr
        }
    }()
    _, err = io.Copy(f, resp.Body)
    return err
}
```

Faça o equivalente em `FetchManifest` (linhas 48–62) para `resp.Body`. Na linha 125, troque o `WriteFile` ignorado por:

```go
if err := os.WriteFile(versionPath, []byte(manifest.Simulador.Version), 0o640); err != nil {
    return "", fmt.Errorf("erro ao gravar versão do simulador: %w", err)
}
```

Não mude URL, hash nem o formato do manifesto. Rode `go test ./internal/release/ ./... -count=1` a partir de `projetos/`. Esse pacote não tem teste hoje; o `go test` do módulo inteiro precisa continuar passando. A saída do golangci-lint deve cair de 18 para 14 issues (as quatro deste arquivo), se nenhuma linha nova gerar `errcheck`.

**Título sugerido:** `refactor(extract function): trata o erro ao fechar recursos em release.go`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Extract Function (Fowler, Refactoring, 2ª ed.)

## Motivação
`release.go` ignora o erro de `Close` em quatro pontos (baseline errcheck, linhas 53, 134, 142 e 152) e o de `WriteFile` na linha 125.

## O que mudou
- Função `closeChecked`.
- `FetchManifest` e `DownloadFile` propagam esse erro quando não há outro.
- `EnsureSimulador` passa a verificar `WriteFile`.

## Como testar
`cd projetos && go test ./... -count=1`
`cd projetos && golangci-lint run ./internal/release/...`
```

**Conflito:** baixo. Único arquivo compartilhado seria `release.go`. A refatoração 6 chama `release.DownloadFile` e `EnsureSimulador`, mas não edita este arquivo. Merge automático se os dois PRs entrarem juntos, desde que a assinatura de `DownloadFile` continue `(dest, url string) (error)` — que é o caso.

---

## 5. Decompose Conditional

**Técnica:** Decompose Conditional (Fowler, *Refactoring*, 2ª ed.).

**Problema / smell:** `Ensure` mistura três decisões no mesmo bloco: JDK já extraído em `~/.hubsaude/jdk`, `java` 21 no `PATH`, e download. Fica difícil ler qual caminho retorna o executável.

**Arquivo e linhas:** `projetos/internal/jdk/jdk.go`, função `Ensure`, linhas 26–44. Não mexa nas funções `extractTarGz` / `extractZip` (é lá que estão os oito `errcheck` deste arquivo).

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/decompose-conditional-jdk-ensure
```

Deixe `Ensure` só com a sequência das três condições:

```go
func (p *Provisioner) Ensure() (string, error) {
    if java := p.localJava(); java != "" {
        return java, nil
    }
    if java := java21OnPath(); java != "" {
        return java, nil
    }
    return p.downloadAndLocate()
}

func java21OnPath() string {
    java, err := exec.LookPath("java")
    if err != nil {
        return ""
    }
    ok, _ := isVersion21(java)
    if !ok {
        return ""
    }
    return java
}

func (p *Provisioner) downloadAndLocate() (string, error) {
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
```

Confira que o texto de log e as mensagens de erro continuam iguais. `localJava` e `isVersion21` permanecem. Não há teste desse pacote; não adicione download real no teste (ele sai para a rede). `go test ./...` deve seguir passando. `go vet ./internal/jdk/` também.

**Título sugerido:** `refactor(decompose conditional): separa a resolução do JDK 21 em Ensure`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Decompose Conditional (Fowler, Refactoring, 2ª ed.)

## Motivação
`Provisioner.Ensure` (jdk.go, linhas 26–44) decide no mesmo bloco se usa o JDK local, o `java` do PATH ou se baixa o Temurin 21.

## O que mudou
- `java21OnPath` e `downloadAndLocate`, extraídas de `Ensure`.
- Sem mudança de comportamento e sem tratar os errcheck de `extractTarGz`/`extractZip`.

## Como testar
`cd projetos && go test ./... -count=1 && go vet ./internal/jdk/`
```

**Conflito:** não fazer em paralelo com outro PR que edite `jdk.go` (por exemplo, limpar os `errcheck` das linhas 82–197). Com as refatorações 1–4 e 6, o arquivo é exclusivo desta.

---

## 6. Extract Function

**Técnica:** Extract Function (Fowler, *Refactoring*, 2ª ed.).

**Problema / smell:** código duplicado. A espera até o processo responder é o mesmo laço em dois comandos:

- `waitReady` em `projetos/cmd/assinatura/cli/server.go`, linhas 168–176
- `waitSimuladorReady` em `projetos/cmd/simulador/cli/lifecycle.go`, linhas 186–194

Os dois fazem `for` + função de prontidão + `time.Sleep(time.Second)` + erro se esgotar as tentativas. O que muda é a condição (`Health` versus `/api/info`) e o texto do erro.

**Arquivos:**

- novo: `projetos/internal/runtime/poll.go`
- novo teste: `projetos/internal/runtime/poll_test.go`
- `projetos/cmd/assinatura/cli/server.go` (só `waitReady`)
- `projetos/cmd/simulador/cli/lifecycle.go` (só `waitSimuladorReady`)

Não unifique `runStart` / `runStop` / `runStatus` neste PR. Os dois ciclos de vida não são iguais (porta, health, download do simulador). Também não retire os `//nolint:errcheck` desses arquivos aqui.

**Passo a passo:**

```bash
git checkout main
git pull --ff-only origin main
git checkout -b refactor/extract-function-poll
```

Crie `projetos/internal/runtime/poll.go`:

```go
package runtime

import (
    "fmt"
    "time"
)

func Poll(attempts int, delay time.Duration, ready func() bool) error {
    for i := 0; i < attempts; i++ {
        if ready() {
            return nil
        }
        time.Sleep(delay)
    }
    return fmt.Errorf("não ficou pronto após %d tentativas", attempts)
}
```

Teste em `poll_test.go` com um contador, sem rede e sem `Sleep` longo (`delay` de um milissegundo, ou uma função que fica pronta na segunda chamada).

Troque o corpo de `waitReady` por uma chamada a `runtime.Poll`, mantendo a mensagem atual se quiser preservar o texto:

```go
func waitReady(inv *invoker.Invoker, maxAttempts int) error {
    err := runtime.Poll(maxAttempts, time.Second, func() bool {
        return inv.Health() == nil
    })
    if err != nil {
        return fmt.Errorf("servidor não ficou pronto após %d tentativas", maxAttempts)
    }
    return nil
}
```

E `waitSimuladorReady`:

```go
func waitSimuladorReady(maxSeconds int) error {
    err := runtime.Poll(maxSeconds, time.Second, simuladorHealthy)
    if err != nil {
        return fmt.Errorf("simulador não ficou pronto após %d segundos", maxSeconds)
    }
    return nil
}
```

`go test ./internal/runtime/ ./cmd/assinatura/... ./cmd/simulador/... -count=1` a partir de `projetos/`.

**Título sugerido:** `refactor(extract function): unifica a espera por prontidão de assinatura e simulador`

**Descrição sugerida:**

```markdown
## Técnica de refatoração
Extract Function (Fowler, Refactoring, 2ª ed.)

## Motivação
`waitReady` (server.go, 168–176) e `waitSimuladorReady` (lifecycle.go, 186–194) repetem o mesmo laço de tentativas.

## O que mudou
- `runtime.Poll`.
- Os dois waits passam a usar essa função.
- Start/stop/status continuam como estão.

## Como testar
`cd projetos && go test ./internal/runtime/ ./cmd/assinatura/... ./cmd/simulador/... -count=1`
```

**Conflito:** nenhum com as refatorações 1–5 se este PR não editar `invoker.go`, `jdk.go`, `release.go` nem `SignatureServer.java`. Não faça em paralelo com outro PR que também altere `server.go` ou `lifecycle.go`.

---

## O que fica de fora de propósito

- PKCS#11 real em `Pkcs11SignatureService.java` (linhas 78–95): o delegate estende `FakeSignatureService` e devolve `"PKCS11_SIM_" + provider.getName()`. Sem HSM no CI. Dívida consciente.
- Os oito `errcheck` de `jdk.go` (extração do tar/zip). Mesma técnica da refatoração 4, noutro PR, depois da 5.
- `install.sh`, `install.ps1` e `release.json`: fluxo antigo do HubSaúde (validador-ui / simulador remoto). Não é refatoração pequena de código.
- README da raiz: ainda descreve a disciplina de 2026-01. Atualizar texto não substitui um PR de refatoração.
