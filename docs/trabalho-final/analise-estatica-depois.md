# Análise estática — depois das refatorações

Registro da etapa 7 do trabalho final. A mesma ferramenta, a mesma versão e o mesmo comando da [baseline](analise-estatica-antes.md) foram repetidos no `main` com as refatorações mergeadas (PRs #1 a #7).

| Item | Valor |
|---|---|
| Ferramenta | golangci-lint |
| Versão | 2.15.0 (binário oficial `golangci-lint-2.15.0-linux-amd64`, compilado com Go 1.27.0) |
| Commit analisado | `78db1b0cc2b9145b370f64d3db7a4054bb748f53` (merge do [PR #7](https://github.com/MatheusBorgez/runner/pull/7), último PR de refatoração em `main`) |
| Data da execução | 2026-11-24 |
| Módulo | `projetos/` (`github.com/kyriosdata/runner`, Go 1.26) |
| Arquivo de configuração | nenhum (linters padrão da v2) |
| Saída bruta (mesmo comando da baseline) | [analise-estatica-depois.txt](analise-estatica-depois.txt) |
| Saída completa (`--max-same-issues=0`) | [analise-estatica-depois-completa.txt](analise-estatica-depois-completa.txt) |

Também foi reexecutado o levantamento completo no commit da tag `v1.0.0`, com o mesmo binário e `--max-same-issues=0`: [analise-estatica-antes-completa.txt](analise-estatica-antes-completa.txt). Esse arquivo não substitui a baseline original; ele só tira o teto padrão do linter, que esconde issues repetidas.

Comando (o mesmo da etapa 3):

```bash
cd projetos
golangci-lint run ./... --timeout 10m
```

`go vet ./...` e `go test ./... -count=1` passaram neste commit (código de saída 0).

## Linters habilitados por padrão

O comando padrão da baseline usa `--max-same-issues=3` (valor default do golangci-lint). Issues com o mesmo texto além da terceira não entram no total. Por isso a saída oficial mostra 18 issues na `v1.0.0` e 17 neste commit, embora o conjunto real seja maior e a queda real seja outra (veja a comparação completa abaixo).

| Linter | Baseline (comando padrão) | Depois (comando padrão) | Baseline completa | Depois completa |
|---|---:|---:|---:|---:|
| errcheck | 18 | 17 | 25 | 21 |
| govet | 0 | 0 | 0 | 0 |
| ineffassign | 0 | 0 | 0 | 0 |
| staticcheck | 0 | 0 | 0 | 0 |
| unused | 0 | 0 | 0 | 0 |
| **Total** | **18** | **17** | **25** | **21** |

O processo terminou com código de saída 1, porque ainda há issues. Nenhum erro de configuração.

## Contagem por arquivo (completa, `--max-same-issues=0`)

Esta é a comparação justa: o mesmo binário 2.15.0, nos dois commits, sem teto de issues repetidas.

| Arquivo | Baseline `v1.0.0` | Depois (`78db1b0`) | Variação |
|---|---:|---:|---:|
| `projetos/internal/jdk/jdk.go` | 10 | 10 | 0 |
| `projetos/internal/release/release.go` | 4 | 0 | −4 |
| `projetos/internal/invoker/invoker_test.go` | 4 | 4 | 0 |
| `projetos/cmd/simulador/cli/lifecycle.go` | 4 | 4 | 0 |
| `projetos/internal/invoker/invoker.go` | 3 | 3 | 0 |
| **Total** | **25** | **21** | **−4** |

As 4 issues eliminadas são as de `release.go`, tratadas no [PR #7](https://github.com/MatheusBorgez/runner/pull/7) (Extract Function: `closeQuietly` e tratamento do `Close` do arquivo gravado em `DownloadFile`). Nenhum outro arquivo teve variação no total. Nenhuma issue nova foi introduzida; o que parece “aparecer” na saída padrão depois da refatoração são apontamentos que já existiam na `v1.0.0` e que o teto `--max-same-issues=3` escondia (por exemplo `resp.Body.Close` em `jdk.go` e em `invoker.go`).

As demais refatorações (Remove Dead Code, Guard Clauses, Decompose Conditional, Extract Function da espera, Replace Magic Number no Java) não foram desenhadas para zerar `errcheck` em `jdk.go`, `invoker.go`, `invoker_test.go` nem `lifecycle.go`.

## Issues restantes (arquivo:linha, lista completa)

| Arquivo:linha | Mensagem |
|---|---|
| `cmd/simulador/cli/lifecycle.go:126` | `resp.Body.Close` sem checagem |
| `cmd/simulador/cli/lifecycle.go:161` | `resp.Body.Close` sem checagem |
| `cmd/simulador/cli/lifecycle.go:173` | `resp.Body.Close` sem checagem |
| `cmd/simulador/cli/lifecycle.go:180` | `conn.Body.Close` sem checagem |
| `internal/invoker/invoker.go:73` | `resp.Body.Close` sem checagem (`Health`) |
| `internal/invoker/invoker.go:89` | `resp.Body.Close` sem checagem (`Shutdown`) |
| `internal/invoker/invoker.go:129` | `resp.Body.Close` sem checagem (`httpPost`) |
| `internal/invoker/invoker_test.go:17` | `Encode` sem checagem |
| `internal/invoker/invoker_test.go:38` | `Encode` sem checagem |
| `internal/invoker/invoker_test.go:59` | `w.Write` sem checagem |
| `internal/invoker/invoker_test.go:78` | `w.Write` sem checagem |
| `internal/jdk/jdk.go:96` | `os.Remove` sem checagem |
| `internal/jdk/jdk.go:128` | `resp.Body.Close` sem checagem |
| `internal/jdk/jdk.go:136` | `f.Close` sem checagem |
| `internal/jdk/jdk.go:146` | `f.Close` sem checagem |
| `internal/jdk/jdk.go:151` | `gz.Close` sem checagem |
| `internal/jdk/jdk.go:175` | `out.Close` sem checagem |
| `internal/jdk/jdk.go:189` | `r.Close` sem checagem |
| `internal/jdk/jdk.go:206` | `rc.Close` sem checagem |
| `internal/jdk/jdk.go:210` | `out.Close` sem checagem |
| `internal/jdk/jdk.go:211` | `rc.Close` sem checagem |

Há outros `Close`/`Remove`/`Kill` com `//nolint:errcheck` (por exemplo em `cmd/assinatura/cli/server.go`). O linter não os conta. Eles continuam sendo dívida da leitura de código, não desta tabela.

O PKCS#11 de fachada em `Pkcs11SignatureService.java` também fica de fora desta análise: o golangci-lint só cobre o módulo Go.
