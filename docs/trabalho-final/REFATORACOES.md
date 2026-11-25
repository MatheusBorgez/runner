# Refatorações realizadas

Registro das refatorações da etapa 5. Cada uma entrou num PR pequeno, com uma técnica do catálogo de Martin Fowler (*Refactoring*, 2ª edição, quando o nome existe lá; o PR #6 usa o nome clássico da 1ª edição). A pipeline `CI` rodou em cada PR (etapa 6). Todas estão mergeadas em `main`.

A ordem dos PRs no GitHub não é a do roteiro inicial: as pessoas trabalharam em paralelo, em arquivos diferentes, e os números seguem a abertura/merge no repositório.

| PR | Técnica | Autor | Arquivos principais | CI |
|---|---|---|---|---|
| [#1](https://github.com/MatheusBorgez/runner/pull/1) | Remove Dead Code | Matheus Andrade Borges | `projetos/assinatura/`, `.github/workflows/ci.yml` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45511730481) |
| [#2](https://github.com/MatheusBorgez/runner/pull/2) | Replace Nested Conditional with Guard Clauses | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | `projetos/internal/invoker/invoker.go` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45512022977) |
| [#3](https://github.com/MatheusBorgez/runner/pull/3) | Pipeline + docs da baseline (não é refatoração) | Matheus Andrade Borges | `.github/workflows/build.yml`, `docs/trabalho-final/` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45512148806) |
| [#4](https://github.com/MatheusBorgez/runner/pull/4) | Decompose Conditional | Matheus Andrade Borges | `projetos/internal/jdk/jdk.go` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45513981877) |
| [#5](https://github.com/MatheusBorgez/runner/pull/5) | Extract Function | Matheus Andrade Borges | `projetos/internal/runtime/poll.go`, `server.go`, `lifecycle.go` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45512814268) |
| [#6](https://github.com/MatheusBorgez/runner/pull/6) | Replace Magic Number with Symbolic Constant | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | `SignatureServer.java` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45513005917) |
| [#7](https://github.com/MatheusBorgez/runner/pull/7) | Extract Function | Gustavo Marçal ([gugualves2002](https://github.com/gugualves2002)) | `projetos/internal/release/release.go` | [verde](https://github.com/MatheusBorgez/runner/actions/runs/45513011008) |

---

## PR #1 — Remove Dead Code

**Técnica:** Remove Dead Code (Fowler, *Refactoring*, 2ª ed.).

**Problema:** `projetos/assinatura/` era um módulo Go antigo (`github.com/kyriosdata/assinatura`) com `main.go`, `cmd/root.go` e `cmd/version.go`. A CLI atual fica em `projetos/cmd/assinatura`. Nada no módulo `github.com/kyriosdata/runner` importava o módulo antigo. O único uso era o workflow legado `.github/workflows/ci.yml`.

**O que mudou:** remoção do diretório `projetos/assinatura/` e do `ci.yml` legado. Sem mudança de comportamento na CLI atual.

**PR:** https://github.com/MatheusBorgez/runner/pull/1

---

## PR #2 — Replace Nested Conditional with Guard Clauses

**Técnica:** Replace Nested Conditional with Guard Clauses (Fowler).

**Problema:** em `projetos/internal/invoker/invoker.go`, `runLocal` chamava `exitErr.ExitCode()` mesmo quando o erro não era `*exec.ExitError` (por exemplo, `java` não encontrado). Nesse caso `exitErr` é `nil` e o CLI dava panic.

**O que mudou:** o fluxo de erro foi separado em guard clauses (sucesso, erro que não é de saída do processo, saída 1, outras saídas). Erros que não são de saída passam a ser encapsulados com `%w`. Entrou o teste `TestLocalInvokerJavaInexistenteNaoDaPanic`.

**PR:** https://github.com/MatheusBorgez/runner/pull/2

---

## PR #3 — Pipeline CI/CD e docs da baseline

Não é uma técnica de Fowler. Cobre os passos 3, 4 e 6 do enunciado: `build.yml` em todo push e em todo PR para `main` (testes Go em Linux/macOS/Windows, build Java, golangci-lint 2.15.0, cross-compile e smoke test) e os documentos da análise estática inicial. O workflow legado `ci.yml` sai neste PR (também saiu no #1; a deleção nos dois lados mescla sem conflito).

**PR:** https://github.com/MatheusBorgez/runner/pull/3

A justificativa de Continuous Delivery (tag `v*`) versus Continuous Deployment está no [RELATORIO.md](RELATORIO.md), seção 5.

---

## PR #4 — Decompose Conditional

**Técnica:** Decompose Conditional (Fowler, *Refactoring*, 2ª ed.).

**Problema:** `Provisioner.Ensure` (`projetos/internal/jdk/jdk.go`) decidia no mesmo bloco se usa o JDK já extraído em `~/.hubsaude/jdk`, o `java` 21 do `PATH` ou se baixa o Temurin 21.

**O que mudou:** `java21OnPath` e `downloadAndLocate` foram extraídas de `Ensure`, que agora só encadeia as três decisões. Entraram testes de JDK local e de versão do `java` no `PATH`, sem download. Sem mudança de comportamento (logs e mensagens de erro iguais). Os `errcheck` de `extractTarGz`/`extractZip` não foram tratados neste PR, de propósito, para não misturar com a técnica da decomposição.

O primeiro `CI` do PR falhou no Windows (https://github.com/MatheusBorgez/runner/actions/runs/45512808359): os testes de `java21OnPath` usavam um script `sh` como `java` falso. O commit seguinte pula esses testes no Windows. A execução seguinte ficou verde.

**PR:** https://github.com/MatheusBorgez/runner/pull/4

---

## PR #5 — Extract Function (`runtime.Poll`)

**Técnica:** Extract Function (Fowler, *Refactoring*, 2ª ed.).

**Problema:** `waitReady` (`cmd/assinatura/cli/server.go`) e `waitSimuladorReady` (`cmd/simulador/cli/lifecycle.go`) repetiam o mesmo laço de tentativas. Mudavam só a condição de prontidão e o texto do erro.

**O que mudou:** nova função `runtime.Poll` (`internal/runtime/poll.go`), com testes de sucesso na segunda tentativa, esgotamento e zero tentativas. Os dois waits passam a usar essa função, com as mesmas mensagens de erro. Start, stop e status continuam como estavam.

**PR:** https://github.com/MatheusBorgez/runner/pull/5

---

## PR #6 — Replace Magic Number with Symbolic Constant

**Técnica:** Replace Magic Number with Symbolic Constant (Fowler, 1ª ed.).

**Problema:** em `projetos/assinador-java/.../http/SignatureServer.java`, os status HTTP `200`, `405` e `422` apareciam como literais espalhados pelos handlers (`handleSign`, `handleValidate`, `handleHealth`, `handleShutdown`).

**O que mudou:** constantes `HTTP_OK`, `HTTP_METHOD_NOT_ALLOWED` e `HTTP_UNPROCESSABLE_ENTITY`. Todas as chamadas de `sendResponse` passam a usá-las. Os valores enviados são os mesmos.

**PR:** https://github.com/MatheusBorgez/runner/pull/6

---

## PR #7 — Extract Function (`closeQuietly` em `release.go`)

**Técnica:** Extract Function (Fowler).

**Problema:** em `projetos/internal/release/release.go`, quatro chamadas `defer x.Close()` descartavam o erro de forma implícita e geravam 4 apontamentos `errcheck`. Uma delas fechava o arquivo recém-gravado em `DownloadFile`, onde o erro do `Close` pode indicar que a gravação no disco falhou.

**O que mudou:** nova função `closeQuietly(io.Closer)`, que torna explícito o descarte do erro em recursos de leitura (`resp.Body` em `FetchManifest` e `DownloadFile`, arquivo aberto em `SHA256File`). Em `DownloadFile`, o arquivo gravado agora devolve o erro do `Close` ao chamador. Se o `io.Copy` falhar, o arquivo é fechado e o erro da cópia é devolvido.

Este PR zera as 4 issues `errcheck` de `release.go` (25 → 21 na contagem completa da análise estática; ver [analise-estatica-depois.md](analise-estatica-depois.md)).

**PR:** https://github.com/MatheusBorgez/runner/pull/7

---

## O que ficou de fora de propósito

- PKCS#11 real em `Pkcs11SignatureService.java`: o delegate estende `FakeSignatureService` e devolve `"PKCS11_SIM_" + provider.getName()`. Sem HSM no CI. Dívida consciente.
- Os `errcheck` que restam em `jdk.go` (extração do tar/zip e download), `invoker.go`, `invoker_test.go` e `lifecycle.go`. Mesma família de problema do PR #7, noutro recorte.
- `install.sh`, `install.ps1` e `release.json`: fluxo antigo do HubSaúde (validador-ui / simulador remoto). Não é refatoração pequena de código.
- README da raiz: ainda descreve a disciplina de 2026-01. Atualizar texto não substitui um PR de refatoração.
