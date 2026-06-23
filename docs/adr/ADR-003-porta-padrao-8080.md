# ADR-003 — Porta padrão do assinador.jar: 8080

**Data:** 2026-01  
**Status:** Aceito  
**Contexto:** US-01.5, US-01.6, US-01.8

## Contexto

O `assinador.jar` em modo servidor precisa de uma porta padrão conhecida para que o CLI possa se conectar sem configuração explícita.

## Decisão

**Porta padrão: 8080.** Configurável via `--port`.

## Justificativa

- 8080 é a porta HTTP alternativa mais comum e raramente usada por serviços de sistema.
- Não requer privilégios de root (portas < 1024 requerem).
- Diferente da porta do Simulador (8443), evitando conflito no mesmo ambiente.

## Portas em uso no Sistema Runner

| Componente | Porta padrão | Protocolo |
|------------|-------------|-----------|
| assinador.jar | 8080 | HTTP |
| simulador.jar | 8443 | HTTPS |

## Consequências

- `assinatura start` sem `--port` inicia na 8080.
- `assinatura sign` sem `--port` procura servidor na 8080.
- Quando a 8080 está ocupada, o usuário especifica `--port` explicitamente.
- Mensagem de erro clara quando porta está tomada por outro processo.
