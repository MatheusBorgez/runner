# Relatório — Implantação e Entrega de Software (2026/2)

- Disciplina: Implantação e Entrega de Software – INF/UFG
- Professora: Sofia Costa Paiva
- Grupo: [NOMES DO GRUPO]
- Repositório: https://github.com/MatheusBorgez/runner
- Baseline (etapa 1): tag `v1.0.0` no commit `8777756d20658998fd0574017b3e09ed9f7a6857`
- Versão final (etapa 8): tag `v2.0.0` [A PREENCHER]

Este texto é um rascunho. Onde aparecer `[A PREENCHER]` ou `[PRINT: ...]`, o grupo cola a evidência antes de entregar.

## 1. Visão do projeto

O Runner é um sistema de linha de comando (fork de `kyriosdata/runner`). Há duas CLIs em Go e um servidor Java:

- `assinatura` sobe, consulta e encerra o `assinador.jar` (`start` / `status` / `stop`) e executa `sign` / `validate`. O modo padrão é HTTP. `--local` chama o JAR num subprocesso.
- `simulador` controla o ciclo de vida do `simulador.jar`.
- O `assinador.jar` é um projeto Maven, Java 21, com endpoints `/health`, `/sign`, `/validate` e `/shutdown`.
- Estado e JAR ficam em `~/.hubsaude/`. Se não houver Java 21 no `PATH` nem em `~/.hubsaude/jdk`, `internal/jdk` baixa um Temurin 21.

Stack: Go 1.26 (`projetos/go.mod`) e Java 21 / Maven (`projetos/assinador-java`). Não há serviço web de produção: a “subida da aplicação” é o servidor local do assinador.

[PRINT: tela do repositório no GitHub, com a URL visível]

## 2. Passo 1 — Repositório e tag 1.0.0

O repositório já era público: https://github.com/MatheusBorgez/runner. A tag anterior, `v0.0.6`, apontava para um commit antigo (`c977b56`) e a Release correspondente não trazia `simulador` nem `assinador.jar`.

A tag anotada `v1.0.0` foi criada no commit atual de `main`:

- SHA: `8777756d20658998fd0574017b3e09ed9f7a6857`
- Mensagem: `fix: remove import de os não utilizado em state_test.go`
- Mensagem da tag: `Release 1.0.0 — baseline do projeto final (Implantação e Entrega de Software 2026/2)`

O push da tag disparou `.github/workflows/release.yml` (gatilho `v*`):

| Run | Resultado | URL |
|---|---|---|
| 45509786237 | falhou | https://github.com/MatheusBorgez/runner/actions/runs/45509786237 |
| 45510218413 | sucesso | https://github.com/MatheusBorgez/runner/actions/runs/45510218413 |

A primeira execução falhou só no job `Testes Java`: o Maven Central respondeu HTTP 429 ao baixar o `maven-clean-plugin`. Os testes Go (Linux, Windows e macOS) passaram. A execução seguinte concluiu testes, build e publicação.

Release publicada: https://github.com/MatheusBorgez/runner/releases/tag/v1.0.0

Artefatos: `assinatura` e `simulador` (linux, windows, darwin, amd64), `assinador-v1.0.0.jar`, `checksums.txt` e assinaturas Cosign (`.pem` / `.sig`).

[PRINT: página da tag `v1.0.0` apontando para `8777756`]

[PRINT: Release `v1.0.0` com a lista de assets]

[PRINT: Actions do run `45510218413` verde]

Acesso da professora e dos outros integrantes: [A PREENCHER]. O repositório é público, então a leitura não depende de convite. O enunciado pede acesso do grupo; o convite em Settings → Collaborators fica registrado no print abaixo.

[PRINT: Settings → Collaborators, com o convite da professora e dos colegas]

## 3. Passo 2 — Leitura de código e dívida técnica

**Técnica:** leitura por abstração stepwise (do detalhe para a responsabilidade do bloco), com um checklist curto: erro ignorado, código morto, duplicação, teste ausente, comportamento que o nome não entrega.

**Trecho lido** (caminho de `assinatura start`):

1. `projetos/cmd/assinatura/cli/server.go` — `runStart` (linhas 57–108), `runStop` (110–136), `runStatus` (138–166), `waitReady` (168–176).
2. `projetos/internal/jdk/jdk.go` — `Ensure` (linhas 26–44) e o download/extração.
3. `projetos/internal/invoker/invoker.go` — `Health` (64–78), `Shutdown` (80–91), `runLocal` (93–113).
4. `projetos/assinador-java/src/main/java/com/kyriosdata/assinador/http/SignatureServer.java` — `start` e `/health` (linhas 61–78 e 124–127).
5. `projetos/assinador-java/src/main/java/com/kyriosdata/assinador/pkcs11/Pkcs11SignatureService.java` — delegate (linhas 78–95).

O bloco faz o seguinte: `runStart` reaproveita um processo vivo se `/health` responde; senão apaga o estado, garante um Java 21, exige `~/.hubsaude/assinador.jar` e sobe `java -jar … server --port`. O servidor Java abre `/health` com HTTP 200 e corpo `{"status":"ok"}`. `sign` e `validate` usam esse HTTP ou, no modo local, um subprocesso.

### Dívidas encontradas

| # | Dívida | Onde | Gravidade | No conjunto de PRs? |
|---|---|---|---|---|
| 1 | Módulo morto `github.com/kyriosdata/assinatura` | `projetos/assinatura/` (`main.go`, `cmd/root.go`, `cmd/version.go`, `go.mod`) | média | sim, refatoração 1 |
| 2 | Workflow que ainda construía esse módulo, só manual | `.github/workflows/ci.yml` (removido na branch da pipeline) | média | tratado na pipeline, não num PR de refatoração |
| 3 | Panic se `cmd.Run` não devolve `*exec.ExitError` | `projetos/internal/invoker/invoker.go`, linhas 104–110 | alta | sim, refatoração 2 |
| 4 | Números mágicos de status HTTP | `SignatureServer.java`, linhas 105, 110, 115, 120, 126, 131, 134 | baixa | sim, refatoração 3 |
| 5 | Erro de `Close` e de `WriteFile` ignorado | `projetos/internal/release/release.go`, linhas 53, 125, 134, 142, 152 | média | sim, refatoração 4 |
| 6 | `Ensure` mistura JDK local, PATH e download | `projetos/internal/jdk/jdk.go`, linhas 26–44 | baixa | sim, refatoração 5 |
| 7 | Laço de espera duplicado | `server.go` 168–176 e `lifecycle.go` 186–194 | baixa | sim, refatoração 6 |
| 8 | `//nolint:errcheck` em `Remove` e `Kill` | `server.go` linhas 73, 129, 133, 153; `lifecycle.go` linhas 68, 131, 135, 149; `sign.go` linha 76 | baixa | não neste conjunto (colidiria com a refatoração 6 se fosse no mesmo arquivo agora) |
| 9 | Oito `errcheck` na extração do JDK | `jdk.go` linhas 82, 122, 137, 161, 175, 192, 196, 197 | média | não; fazer depois da refatoração 5, com a mesma técnica da 4 |
| 10 | PKCS#11 de fachada | `Pkcs11SignatureService.java` linhas 78–95: `Pkcs11DelegateService` estende `FakeSignatureService` e devolve `"PKCS11_SIM_" + provider.getName() + "_BASE64=="`. O `KeyStore.getInstance("PKCS11", provider)` está comentado na linha 92 | alta como dívida de produto; fora de escopo | não. Sem HSM. Dívida consciente |
| 11 | Ciclo de vida de `assinatura` e `simulador` parecidos, mas não iguais (porta, health, download) | `server.go` e `cmd/simulador/cli/lifecycle.go` | média | só o laço de espera (refatoração 6). Unificar start/stop inteiro arrisca comportamento |
| 12 | Pacotes sem teste | `internal/jdk`, `internal/release`, `cmd/simulador/cli`, a maior parte de `cmd/assinatura/cli` | média | a refatoração 2 e a 6 acrescentam teste pontual; o resto fica de fora |
| 13 | Instaladores e manifesto do fluxo HubSaúde antigo | `install.sh`, `install.ps1`, `release.json` | baixa para este trabalho | não |
| 14 | README da raiz descreve a disciplina de 2026-01, não este projeto final | `README.md` | baixa | não substitui refatoração de código |

[PRINT: trecho de `invoker.go` 104–110 aberto no editor, com a linha do `ExitCode` visível]

[PRINT: trecho de `Pkcs11SignatureService.java` 87–95]

## 4. Passo 3 — Análise estática (baseline)

**Ferramenta:** golangci-lint **2.15.0**, sem arquivo de configuração (linters padrão da linha 2: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`).

Comando, no commit da tag `v1.0.0`:

```bash
cd projetos
golangci-lint run ./... --timeout 10m
```

Saída bruta: [analise-estatica-antes.txt](analise-estatica-antes.txt). Resumo: [analise-estatica-antes.md](analise-estatica-antes.md).

| Linter | Issues |
|---|---|
| errcheck | 18 |
| govet | 0 |
| ineffassign | 0 |
| staticcheck | 0 |
| unused | 0 |
| **Total** | **18** |

Por arquivo: `internal/jdk/jdk.go` 8, `internal/release/release.go` 4, `internal/invoker/invoker_test.go` 4, `internal/invoker/invoker.go` 1, `cmd/simulador/cli/lifecycle.go` 1.

O comando não analisa `projetos/assinatura` (módulo Go separado). Código de saída do linter: 1.

`staticcheck` e `govet` não reportaram o panic de `exitErr` (dívida 3). A baseline não substitui a leitura de código.

[PRINT: trecho final da saída, com a linha `18 issues` e `errcheck: 18`]

## 5. Passo 4 — Pipeline de CI/CD

Arquivo: `.github/workflows/build.yml` (nome do workflow: `CI`).

Gatilhos:

- `push` em qualquer branch (filtros `*` e `**`; tag não dispara este arquivo, para não competir com `release.yml`);
- `pull_request` com base `main`.

Isso cobre commit em branch e cada PR para `main`.

| Job | O que faz |
|---|---|
| `test-go` | `go vet` e `go test ./...` em `projetos/`, no Ubuntu, Windows e macOS, Go 1.26 |
| `lint-go` | golangci-lint **2.15.0** (mesmo binário da baseline). Grava o relatório no log, no sumário do job e no artefato `golangci-lint`. `--issues-exit-code=0`: as 18 issues já existentes não quebram o build |
| `build-go` | cross-compile de `assinatura` e `simulador` (linux, windows, darwin) |
| `build-java` | `mvn -B clean verify` com Java 21 Temurin. Até três tentativas, porque o Maven Central respondeu 429 na primeira Release `v1.0.0` |
| `smoke` | depende de `test-go` e `build-java`. Baixa o JAR produzido, copia para `~/.hubsaude/assinador.jar`, compila a CLI, executa `assinatura start`, `GET /health`, `assinatura sign`, `assinatura status` e `assinatura stop`. Um passo com `if: always()` chama `stop` de novo |

O job de lint não usa `--new-from-rev`. Esse filtro reapareceria issues antigas em qualquer linha que um PR de refatoração tocasse, e o build ficaria vermelho sem código novo defeituoso. O relatório continua sendo gerado em todo push e em todo PR. Depois que a contagem cair, o grupo pode tirar `--issues-exit-code=0` e passar a falhar o job.

`.github/workflows/ci.yml` foi removido nesta branch. Ele só rodava manualmente, compilava o módulo morto `projetos/assinatura` com Go 1.25 e ainda podia publicar release. A pipeline da disciplina é o `build.yml`. A publicação de versão continua em `.github/workflows/release.yml`, no push de uma tag `v*`.

```text
push ou pull_request
        │
        ├─ test-go (3 SOs) ──► build-go
        │
        ├─ lint-go (relatório, não bloqueia)
        │
        └─ build-java (mvn verify) ──► smoke
                                      start → GET /health → sign → stop
```

Verificação nesta branch, antes do PR: o run https://github.com/MatheusBorgez/runner/actions/runs/45511643280 concluiu com sucesso (testes Go nos três sistemas, Java, golangci-lint, cross-compile e smoke). Um run anterior, https://github.com/MatheusBorgez/runner/actions/runs/45511315146, falhou só no smoke: `go build -o assinatura` gravou o binário dentro do diretório `projetos/assinatura`. O commit seguinte gera o binário em `RUNNER_TEMP`. Enquanto este workflow não estiver em `main`, a lista de Actions ainda mostra o nome antigo **Build & Test**; o arquivo já se chama `CI`.

[PRINT: grafo dos jobs na aba Actions, com `smoke` verde]

[PRINT: log do smoke mostrando `assinador.jar iniciado` e o JSON de `/health`]

### Entrega contínua, e não implantação contínua

A pipeline de **integração contínua** roda em todo push de branch e em todo PR para `main`: testes Go, testes Java, análise estática e a subida da aplicação num runner efêmero do GitHub Actions (`start`, `/health`, `sign`, `stop`). O runner some no fim do job. Não há ambiente de produção.

A pipeline de **entrega contínua** é o `release.yml`. Ela gera binários, o JAR, checksums e assinatura Cosign e publica uma GitHub Release **somente quando alguém cria uma tag `v*`**. A tag é a decisão humana de publicar. Foi o que ocorreu na `v1.0.0`.

Não se usa **implantação contínua** (deploy automático em produção a cada commit). O produto é uma CLI mais um servidor local, iniciado na máquina de quem o usa. Não existe servidor de produção para receber esse deploy, e um commit intermediário não deve ser distribuído como a versão instalada pelos usuários. A versão instalável é a Release da tag.

## 6. Passos 5 e 6 — Refatorações em PRs, com a pipeline em cada um

As refatorações **não foram feitas** nesta branch. O roteiro, com técnica, arquivos, comandos e título de PR, está em [REFATORACOES.md](REFATORACOES.md). São seis PRs, em arquivos disjuntos, para duas pessoas trabalharem em paralelo:

1. Remove Dead Code — `projetos/assinatura/`
2. Replace Nested Conditional with Guard Clauses — `invoker.go`
3. Replace Magic Number with Symbolic Constant — `SignatureServer.java`
4. Extract Function — `release.go`
5. Decompose Conditional — `jdk.Ensure`
6. Extract Function — espera compartilhada (`runtime.Poll`)

Cada PR deve usar a branch da pipeline já mergeada em `main`, para o workflow `CI` aparecer nos checks.

| PR | Técnica | URL | CI |
|---|---|---|---|
| [A PREENCHER] | Remove Dead Code | [A PREENCHER] | [A PREENCHER] |
| [A PREENCHER] | Replace Nested Conditional with Guard Clauses | [A PREENCHER] | [A PREENCHER] |
| [A PREENCHER] | Replace Magic Number with Symbolic Constant | [A PREENCHER] | [A PREENCHER] |
| [A PREENCHER] | Extract Function (`release.go`) | [A PREENCHER] | [A PREENCHER] |
| [A PREENCHER] | Decompose Conditional | [A PREENCHER] | [A PREENCHER] |
| [A PREENCHER] | Extract Function (`Poll`) | [A PREENCHER] | [A PREENCHER] |

[PRINT: aba Checks de cada PR, com o workflow `CI` verde, inclusive o job `smoke`]

## 7. Passo 7 — Análise estática outra vez

Rodar de novo, no `main` já com as refatorações mergeadas, **a mesma versão e o mesmo comando** da etapa 3:

```bash
cd projetos
golangci-lint version   # tem que ser 2.15.0
golangci-lint run ./... --timeout 10m | tee docs/trabalho-final/analise-estatica-depois.txt
```

| Métrica | Baseline `v1.0.0` | Depois | Variação |
|---|---|---|---|
| errcheck | 18 | [A PREENCHER] | [A PREENCHER] |
| govet | 0 | [A PREENCHER] | [A PREENCHER] |
| ineffassign | 0 | [A PREENCHER] | [A PREENCHER] |
| staticcheck | 0 | [A PREENCHER] | [A PREENCHER] |
| unused | 0 | [A PREENCHER] | [A PREENCHER] |
| Total | 18 | [A PREENCHER] | [A PREENCHER] |

A refatoração 4, se feita como está no roteiro, deve eliminar as 4 issues de `release.go`. As outras refatorações não foram desenhadas para zerar `jdk.go` nem os testes do invoker. O texto da comparação precisa dizer o que melhorou e o que ficou (PKCS#11 simulado, `errcheck` do JDK, instaladores antigos).

[A PREENCHER: parágrafo da comparação]

[PRINT: as duas saídas lado a lado, ou uma tabela preenchida]

[PRINT: gráfico de barras antes/depois, se o grupo fizer]

## 8. Passo 8 — Tag 2.0.0

Criar a tag **só depois** que os PRs de refatoração estiverem mergeados em `main` e a comparação da etapa 7 estiver salva. Não criar esta tag a partir da branch da pipeline.

```bash
git checkout main
git pull --ff-only origin main
git tag -a v2.0.0 -m "Release 2.0.0 — após refatorações e pipeline CI/CD do projeto final"
git push origin v2.0.0
```

O `release.yml` publica a Release `v2.0.0` (binários, JAR, checksums, Cosign). Conferir o Actions e a página da Release.

[PRINT: tag `v2.0.0`]

[PRINT: Release `v2.0.0`]

[PRINT: Actions do `release.yml` verde nessa tag]

SHA do commit da tag: [A PREENCHER]

## 9. Conclusão

[A PREENCHER: o que a baseline mostrou, o que a pipeline passa a garantir em todo PR, e o que o grupo escolheu não refatorar]

## Referências

- Martin Fowler. *Refactoring: Improving the Design of Existing Code*. 2ª ed. Addison-Wesley, 2018. (1ª ed., 1999, para Replace Magic Number with Symbolic Constant.)
- Repositório: https://github.com/MatheusBorgez/runner
- golangci-lint 2.15.0: https://github.com/golangci/golangci-lint/releases/tag/v2.15.0
