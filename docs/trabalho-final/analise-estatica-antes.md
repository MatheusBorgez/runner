# Análise estática — baseline (antes das refatorações)

Registro da etapa 3 do trabalho final. A mesma ferramenta, a mesma versão e o mesmo comando devem ser repetidos na etapa 7, depois das refatorações.

| Item | Valor |
|---|---|
| Ferramenta | golangci-lint |
| Versão | 2.15.0 (binário oficial `golangci-lint-2.15.0-linux-amd64`, compilado com Go 1.27.0) |
| Commit | `8777756d20658998fd0574017b3e09ed9f7a6857` |
| Tag | `v1.0.0` |
| Data da execução | 2026-11-24 |
| Módulo | `projetos/` (`github.com/kyriosdata/runner`, Go 1.26) |
| Arquivo de configuração | nenhum (linters padrão da v2) |
| Saída bruta | [analise-estatica-antes.txt](analise-estatica-antes.txt) |

Comando:

```bash
cd projetos
golangci-lint run ./... --timeout 10m
```

`./...` não entra em `projetos/assinatura/`: aquele diretório é outro módulo Go (`github.com/kyriosdata/assinatura`). A dívida dele aparece na leitura de código, não nesta contagem.

## Linters habilitados por padrão

| Linter | Issues |
|---|---|
| errcheck | 18 |
| govet | 0 |
| ineffassign | 0 |
| staticcheck | 0 |
| unused | 0 |
| **Total** | **18** |

O processo terminou com código de saída 1, porque havia issues. Nenhum erro de configuração.

## Contagem por arquivo

| Arquivo | Issues (todas `errcheck`) |
|---|---|
| `projetos/internal/jdk/jdk.go` | 8 |
| `projetos/internal/release/release.go` | 4 |
| `projetos/internal/invoker/invoker_test.go` | 4 |
| `projetos/internal/invoker/invoker.go` | 1 |
| `projetos/cmd/simulador/cli/lifecycle.go` | 1 |
| **Total** | **18** |

## Issues (arquivo:linha)

| Arquivo:linha | Mensagem |
|---|---|
| `cmd/simulador/cli/lifecycle.go:180` | `conn.Body.Close` sem checagem |
| `internal/invoker/invoker.go:73` | `resp.Body.Close` sem checagem |
| `internal/invoker/invoker_test.go:15` | `Encode` sem checagem |
| `internal/invoker/invoker_test.go:36` | `Encode` sem checagem |
| `internal/invoker/invoker_test.go:57` | `w.Write` sem checagem |
| `internal/invoker/invoker_test.go:76` | `w.Write` sem checagem |
| `internal/jdk/jdk.go:82` | `os.Remove` sem checagem |
| `internal/jdk/jdk.go:122` | `f.Close` sem checagem |
| `internal/jdk/jdk.go:137` | `gz.Close` sem checagem |
| `internal/jdk/jdk.go:161` | `out.Close` sem checagem |
| `internal/jdk/jdk.go:175` | `r.Close` sem checagem |
| `internal/jdk/jdk.go:192` | `rc.Close` sem checagem |
| `internal/jdk/jdk.go:196` | `out.Close` sem checagem |
| `internal/jdk/jdk.go:197` | `rc.Close` sem checagem |
| `internal/release/release.go:53` | `resp.Body.Close` sem checagem |
| `internal/release/release.go:134` | `resp.Body.Close` sem checagem |
| `internal/release/release.go:142` | `f.Close` sem checagem |
| `internal/release/release.go:152` | `f.Close` sem checagem |

Há outros `Close`/`Remove` com `//nolint:errcheck` (por exemplo em `cmd/assinatura/cli/server.go`). O linter não os conta. Eles continuam sendo dívida da leitura de código, não desta tabela.

O defeito de ponteiro nulo em `internal/invoker/invoker.go` (por volta das linhas 104–110) também não aparece nesta baseline: `staticcheck` e `govet` não o apontaram. Ele fica registrado na leitura de código e na refatoração 2.

A reexecução da etapa 7 mostrou que o comando padrão esconde issues repetidas (`--max-same-issues=3`). Os 18 apontamentos acima são a saída oficial desta baseline e não foram alterados. A lista completa no mesmo commit, com o teto desligado, está em [analise-estatica-antes-completa.txt](analise-estatica-antes-completa.txt) (25 issues `errcheck`) e é usada na comparação da [análise de depois](analise-estatica-depois.md).
