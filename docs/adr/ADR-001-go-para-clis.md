# ADR-001 — Go para CLIs multiplataforma

**Status:** Aceito | **Contexto:** US-01, US-03, US-05

## Decisão

Usar **Go 1.26** para os CLIs `assinatura` e `simulador`.

## Justificativa

- Cross-compilation nativa (`GOOS`/`GOARCH`): binários estáticos para Windows/Linux/macOS a partir de um único runner no CI, sem runtime no cliente.
- Biblioteca padrão rica: HTTP, processos externos, tar/zip — sem dependências externas.
- Cobra é o framework de CLI mais adotado no ecossistema Go.

## Alternativas rejeitadas

- **Python/Node.js**: requerem runtime instalado; distribuição complexa.
- **Rust**: binários igualmente estáticos, mas curva de aprendizado desproporcional para o contexto.

## Consequências

Dois binários em `cmd/assinatura` e `cmd/simulador` no mesmo módulo `github.com/kyriosdata/runner`, compartilhando `internal/`.
