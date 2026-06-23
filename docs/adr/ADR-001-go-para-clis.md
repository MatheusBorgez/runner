# ADR-001 — Go para CLIs multiplataforma

**Data:** 2026-01  
**Status:** Aceito  
**Contexto:** US-01, US-03, US-05

## Contexto

O Sistema Runner precisa de dois CLIs distribuíveis (`assinatura` e `simulador`) para Windows, Linux e macOS. Os usuários não devem precisar instalar nenhum runtime para executar os CLIs.

## Decisão

Usar **Go 1.26** para os dois CLIs.

## Justificativa

- Cross-compilation nativa (`GOOS`/`GOARCH`): gera binários estáticos para as 3 plataformas a partir de um único runner Linux no CI.
- Biblioteca padrão rica: HTTP client/server, processos externos, manipulação de arquivos, extração de tar/zip — tudo sem dependências externas.
- Binário único sem runtime: o usuário baixa e executa imediatamente.
- Cobra (CLI framework) bem estabelecido para estrutura de subcomandos.

## Alternativas consideradas

- **Python**: requer runtime instalado; distribuição complexa (PyInstaller gera binários grandes e lentos).
- **Node.js/Deno**: similar ao Python em termos de distribuição.
- **Rust**: binários igualmente estáticos, mas curva de aprendizado maior e toolchain mais complexo para o contexto da disciplina.

## Consequências

- Dois binários no mesmo módulo Go (`github.com/kyriosdata/runner`) em `cmd/assinatura` e `cmd/simulador`, compartilhando `internal/`.
- CI usa `actions/setup-go@v5` com `go-version: "1.26"`.
