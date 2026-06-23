# ADR-003 — Porta padrão do assinador.jar: 8080

**Status:** Aceito | **Contexto:** US-01.5, US-01.6, US-01.8

## Decisão

Porta padrão **8080**, configurável via `--port`.

## Justificativa

- Não requer privilégios de root (< 1024 exigem).
- Não conflita com o Simulador (8443).
- É a porta HTTP alternativa mais reconhecida.

| Componente    | Porta | Protocolo |
|---------------|-------|-----------|
| assinador.jar | 8080  | HTTP      |
| simulador.jar | 8443  | HTTPS     |
