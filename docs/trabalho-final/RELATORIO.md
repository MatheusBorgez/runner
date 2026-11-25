# Relatório — Implantação e Entrega de Software (2026/2)

- Disciplina: Implantação e Entrega de Software – INF/UFG
- Professora: Sofia Costa Paiva
- Grupo: Matheus Andrade Borges; Gustavo Marçal (GitHub: [gugualves2002](https://github.com/gugualves2002)) [A PREENCHER: demais integrantes, se houver]
- Repositório: https://github.com/MatheusBorgez/runner
- Baseline (etapa 1): tag `v1.0.0` no commit `8777756d20658998fd0574017b3e09ed9f7a6857`
- Versão final (etapa 8): tag `v2.0.0` (criada no `main` depois deste relatório; ver https://github.com/MatheusBorgez/runner/releases/tag/v2.0.0)

Os trechos técnicos (PRs, pipeline, análise estática e justificativa de Continuous Delivery) já estão preenchidos. Onde aparecer `[A PREENCHER]` ou `[PRINT: ...]`, falta evidência que só o grupo tem (prints de tela, convites no GitHub, demais nomes se houver).

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

`.github/workflows/ci.yml` foi removido em `main` (PRs #1 e #3). Ele só rodava manualmente, compilava o módulo morto `projetos/assinatura` com Go 1.25 e ainda podia publicar release. A pipeline da disciplina é o `build.yml`. A publicação de versão continua em `.github/workflows/release.yml`, no push de uma tag `v*`.

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

O workflow entrou em `main` pelo [PR #3](https://github.com/MatheusBorgez/runner/pull/3) (CI verde: https://github.com/MatheusBorgez/runner/actions/runs/45512148806). A verificação na branch, antes desse PR, foi o run https://github.com/MatheusBorgez/runner/actions/runs/45511643280 (testes Go nos três sistemas, Java, golangci-lint, cross-compile e smoke). Um run anterior, https://github.com/MatheusBorgez/runner/actions/runs/45511315146, falhou só no smoke: `go build -o assinatura` gravou o binário dentro do diretório `projetos/assinatura`. O commit seguinte gera o binário em `RUNNER_TEMP`. Depois do merge das refatorações, o `CI` em `main` (merge do PR #7) também passou: https://github.com/MatheusBorgez/runner/actions/runs/45514491365.

[PRINT: grafo dos jobs na aba Actions, com `smoke` verde]

[PRINT: log do smoke mostrando `assinador.jar iniciado` e o JSON de `/health`]

### Entrega contínua, e não implantação contínua

A pipeline de **integração contínua** roda em todo push de branch e em todo PR para `main`: testes Go, testes Java, análise estática e a subida da aplicação num runner efêmero do GitHub Actions (`start`, `/health`, `sign`, `stop`). O runner some no fim do job. Não há ambiente de produção.

A pipeline de **entrega contínua** é o `release.yml`. Ela gera binários, o JAR, checksums e assinatura Cosign e publica uma GitHub Release **somente quando alguém cria uma tag `v*`**. A tag é a decisão humana de publicar. Foi o que ocorreu na `v1.0.0`.

Não se usa **implantação contínua** (deploy automático em produção a cada commit). O produto é uma CLI mais um servidor local, iniciado na máquina de quem o usa. Não existe servidor de produção para receber esse deploy, e um commit intermediário não deve ser distribuído como a versão instalada pelos usuários. A versão instalável é a Release da tag.

## 6. Passos 5 e 6 — Refatorações em PRs, com a pipeline em cada um

As seis refatorações de código e o PR da pipeline já foram mergeados em `main`. O detalhe de cada técnica está em [REFATORACOES.md](REFATORACOES.md). A numeração abaixo é a do GitHub (o roteiro original tinha outra ordem: a 3 do roteiro é o PR #6, a 4 é o PR #7, a 5 é o PR #4 e a 6 é o PR #5).

Cada PR de refatoração correu com o workflow `CI` nos checks. O PR #1 foi aberto e mergeado antes do PR #3 entrar em `main`; ainda assim os jobs de teste e build do `build.yml` rodaram no PR. A partir do PR #2, o `CI` completo (incluindo lint e, quando a branch já tinha o workflow, smoke) aparece nos checks.

| PR | Técnica | Autor | URL | CI (run do PR, verde) |
|---|---|---|---|---|
| [#1](https://github.com/MatheusBorgez/runner/pull/1) | Remove Dead Code | Matheus Andrade Borges | https://github.com/MatheusBorgez/runner/pull/1 | https://github.com/MatheusBorgez/runner/actions/runs/45511730481 |
| [#2](https://github.com/MatheusBorgez/runner/pull/2) | Replace Nested Conditional with Guard Clauses | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | https://github.com/MatheusBorgez/runner/pull/2 | https://github.com/MatheusBorgez/runner/actions/runs/45512022977 |
| [#3](https://github.com/MatheusBorgez/runner/pull/3) | Pipeline CI/CD + docs da baseline (não é refatoração de código) | Matheus Andrade Borges | https://github.com/MatheusBorgez/runner/pull/3 | https://github.com/MatheusBorgez/runner/actions/runs/45512148806 |
| [#4](https://github.com/MatheusBorgez/runner/pull/4) | Decompose Conditional (`jdk.Ensure`) | Matheus Andrade Borges | https://github.com/MatheusBorgez/runner/pull/4 | https://github.com/MatheusBorgez/runner/actions/runs/45513981877 |
| [#5](https://github.com/MatheusBorgez/runner/pull/5) | Extract Function (`runtime.Poll`) | Matheus Andrade Borges | https://github.com/MatheusBorgez/runner/pull/5 | https://github.com/MatheusBorgez/runner/actions/runs/45512814268 |
| [#6](https://github.com/MatheusBorgez/runner/pull/6) | Replace Magic Number with Symbolic Constant | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | https://github.com/MatheusBorgez/runner/pull/6 | https://github.com/MatheusBorgez/runner/actions/runs/45513005917 |
| [#7](https://github.com/MatheusBorgez/runner/pull/7) | Extract Function (`closeQuietly` em `release.go`) | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | https://github.com/MatheusBorgez/runner/pull/7 | https://github.com/MatheusBorgez/runner/actions/runs/45513011008 |

O [PR #4](https://github.com/MatheusBorgez/runner/pull/4) teve uma execução vermelha no Windows (https://github.com/MatheusBorgez/runner/actions/runs/45512808359): os testes de `java21OnPath` usavam um script `sh` como `java` falso, o que não roda naquele SO. O commit seguinte pula esses testes no Windows. A execução seguinte ficou verde e o PR foi mergeado.

[PRINT: aba Checks de cada PR, com o workflow `CI` verde, inclusive o job `smoke` quando ele fizer parte do workflow daquele PR]

## 7. Passo 7 — Análise estática outra vez

Foi rodado de novo, no `main` já com as refatorações mergeadas (commit `78db1b0cc2b9145b370f64d3db7a4054bb748f53`, merge do PR #7), **a mesma versão e o mesmo comando** da etapa 3:

```bash
cd projetos
golangci-lint version   # 2.15.0
golangci-lint run ./... --timeout 10m
```

Saída bruta: [analise-estatica-depois.txt](analise-estatica-depois.txt). Resumo: [analise-estatica-depois.md](analise-estatica-depois.md). Neste mesmo commit, `go vet ./...` e `go test ./... -count=1` passaram (código de saída 0).

O comando padrão do golangci-lint limita issues com o mesmo texto a 3 (`--max-same-issues=3`). A baseline original e a reexecução “oficial” portanto mostram 18 e 17 issues, mas escondem apontamentos repetidos de `Close`. A comparação justa, com `--max-same-issues=0` no mesmo binário 2.15.0, está em [analise-estatica-antes-completa.txt](analise-estatica-antes-completa.txt) e [analise-estatica-depois-completa.txt](analise-estatica-depois-completa.txt).

| Métrica | Baseline `v1.0.0` (comando padrão) | Depois (comando padrão) | Baseline completa | Depois completa | Variação real |
|---|---:|---:|---:|---:|---:|
| errcheck | 18 | 17 | 25 | 21 | −4 |
| govet | 0 | 0 | 0 | 0 | 0 |
| ineffassign | 0 | 0 | 0 | 0 | 0 |
| staticcheck | 0 | 0 | 0 | 0 | 0 |
| unused | 0 | 0 | 0 | 0 | 0 |
| **Total** | **18** | **17** | **25** | **21** | **−4** |

Por arquivo, na contagem completa:

| Arquivo | Antes | Depois | Variação |
|---|---:|---:|---:|
| `internal/jdk/jdk.go` | 10 | 10 | 0 |
| `internal/release/release.go` | 4 | 0 | −4 |
| `internal/invoker/invoker_test.go` | 4 | 4 | 0 |
| `cmd/simulador/cli/lifecycle.go` | 4 | 4 | 0 |
| `internal/invoker/invoker.go` | 3 | 3 | 0 |
| **Total** | **25** | **21** | **−4** |

A queda inteira vem do [PR #7](https://github.com/MatheusBorgez/runner/pull/7): as 4 issues `errcheck` de `release.go` saíram (`closeQuietly` nos `Close` de leitura e tratamento do `Close` do arquivo gravado). Os totais dos outros arquivos não mudaram. O que a saída padrão “ganha” em `jdk.go` e `invoker.go` depois da refatoração não é issue nova: é apontamento que já existia na `v1.0.0` e que o teto de 3 ocorrências iguais escondia.

O que ficou de propósito: os `errcheck` da extração do JDK e dos `Close` em `invoker.go`, `invoker_test.go` e `lifecycle.go`; o PKCS#11 de fachada em `Pkcs11SignatureService.java` (sem HSM no CI); e os instaladores / `release.json` do fluxo antigo do HubSaúde. Isso não entra no golangci-lint do módulo Go.

[PRINT: as duas saídas lado a lado, ou uma tabela preenchida — os arquivos `.txt` e a tabela acima já registram os números]

[PRINT: gráfico de barras antes/depois, se o grupo fizer]

## 8. Passo 8 — Tag 2.0.0

A tag anotada `v2.0.0` é criada no `main` **depois** que os PRs de refatoração (#1, #2, #4, #5, #6 e #7), o PR da pipeline (#3) e o PR com esta comparação e este relatório estiverem mergeados.

```bash
git checkout main
git pull --ff-only origin main
git tag -a v2.0.0 -m "Release 2.0.0 — após refatorações e pipeline CI/CD do projeto final"
git push origin v2.0.0
```

O `release.yml` publica a Release `v2.0.0` (binários, JAR, checksums, Cosign). A página da Release e o run do Actions ficam em:

- Tag e Release: https://github.com/MatheusBorgez/runner/releases/tag/v2.0.0
- SHA do commit da tag: o commit de `main` que contém este relatório e a análise de depois (ver a página da tag)

[PRINT: tag `v2.0.0`]

[PRINT: Release `v2.0.0`]

[PRINT: Actions do `release.yml` verde nessa tag]

## 9. Conclusão

A baseline na tag `v1.0.0` (commit `8777756`) mostrou, no comando padrão do golangci-lint 2.15.0, 18 issues `errcheck`. Com o teto de issues repetidas desligado, eram 25, concentradas em `jdk.go`, `release.go`, testes do invoker, `invoker.go` e `lifecycle.go`. A leitura de código apontou ainda o módulo morto `projetos/assinatura/`, o panic em `runLocal` quando o `java` não existia, números mágicos de status HTTP no assinador Java, o condicional carregado de `jdk.Ensure`, a espera de prontidão duplicada e o PKCS#11 apenas simulado.

A pipeline `CI` (PR #3) passa a rodar em todo push de branch e em todo PR para `main`: testes Go nos três sistemas, `go vet`, golangci-lint (relatório, sem quebrar o build pelas issues já conhecidas), build Java, cross-compile e smoke test (`assinatura start`, `/health`, `sign`, `stop`). A publicação de versão continua sendo uma **entrega contínua** disparada por tag `v*`, não uma implantação contínua em produção: o produto é uma CLI com servidor local, e a versão instalável é a GitHub Release da tag.

As seis refatorações entraram em PRs separados, cada um nomeando a técnica de Fowler, com CI verde antes do merge. A reexecução do linter no `main` depois desses PRs mostra 21 issues na contagem completa (25 → 21, −4), todas as 4 saídas em `release.go`. O restante da dívida listada acima foi deixado de propósito: zerar `errcheck` em `jdk.go` ou implementar PKCS#11 de verdade sairia do recorte de PRs pequenos e, no segundo caso, exigiria HSM no CI.

## Referências

- Martin Fowler. *Refactoring: Improving the Design of Existing Code*. 2ª ed. Addison-Wesley, 2018. (1ª ed., 1999, para Replace Magic Number with Symbolic Constant.)
- Repositório: https://github.com/MatheusBorgez/runner
- golangci-lint 2.15.0: https://github.com/golangci/golangci-lint/releases/tag/v2.15.0
