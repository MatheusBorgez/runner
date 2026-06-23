# ADR-002 — Modo servidor como padrão de invocação do assinador.jar

**Status:** Aceito | **Contexto:** US-01.5–01.7, criterios.md E2

## Decisão

O modo servidor (HTTP) é o padrão. O modo local (subprocess) é ativado explicitamente via `--local`.

## Justificativa

`criterios.md` E2 exige isso explicitamente. Além disso, elimina o overhead de cold start da JVM para múltiplas invocações sequenciais.

## Protocolo de descoberta de instância

1. Lê `~/.hubsaude/assinador.state.json` (PID + porta).
2. Verifica se o PID está ativo com `signal(0)`.
3. Faz `GET /health` — health check real, não só TCP.
4. Se passa → reutiliza. Se falha → remove estado corrompido e cai para modo local.
