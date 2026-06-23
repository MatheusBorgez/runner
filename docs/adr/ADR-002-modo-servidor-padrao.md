# ADR-002 — Modo servidor como padrão de invocação do assinador.jar

**Data:** 2026-01  
**Status:** Aceito  
**Contexto:** US-01.5, US-01.6, US-01.7, E2 de criterios.md

## Contexto

O CLI `assinatura` pode invocar o `assinador.jar` de dois modos:
- **Local (subprocess):** `java -jar assinador.jar sign ...` — cold start a cada chamada (~500ms–2s de overhead JVM).
- **Servidor (HTTP):** `assinador.jar server` fica em background, o CLI faz `POST /sign` — warm start.

## Decisão

**O modo servidor é o padrão.** O modo local é explicitamente ativado via flag `--local`.

## Justificativa

- Especificação (`criterios.md` E2): "Modo servidor é o padrão; modo local deve ser explicitamente ativado."
- Menor latência para múltiplas invocações (integradores fazem chamadas repetidas).
- Melhor throughput para scripts de automação.
- O `assinatura start` é idempotente: detecta instância ativa por health check real (não só "porta ocupada").

## Protocolo de descoberta de instância

1. Lê `~/.hubsaude/assinador.state.json` (PID + porta registrados no `start`).
2. Verifica se PID está ativo com `signal(0)`.
3. Faz `GET /health` na porta registrada — health check real, não apenas TCP.
4. Se health check passa → usa servidor existente.
5. Se falha → apaga estado corrompido → cai para modo local.

## Consequências

- Usuários que quiserem execução esporádica usam `--local`.
- PID e porta armazenados em `~/.hubsaude/assinador.state.json`.
- `assinatura start` deve ser rodado antes do primeiro `sign`/`validate` sem `--local`.
