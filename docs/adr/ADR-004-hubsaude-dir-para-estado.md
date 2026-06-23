# ADR-004 — Diretório ~/.hubsaude/ para estado e artefatos

**Data:** 2026-01  
**Status:** Aceito  
**Contexto:** US-01.5, US-03, US-04, plano-preliminar.md

## Contexto

O Sistema Runner precisa persistir estado entre invocações (PID, porta, versão do JAR) e armazenar artefatos baixados (JRE/JDK, simulador.jar, assinador.jar).

## Decisão

Usar `~/.hubsaude/` como diretório base para todo o estado local do Sistema Runner.

## Estrutura

```
~/.hubsaude/
├── assinador.jar          — JAR do assinador (copiado aqui na instalação)
├── simulador.jar          — JAR do simulador (baixado dinamicamente)
├── simulador.version      — versão do simulador instalado
├── assinador.state.json   — PID e porta do assinador em execução
├── simulador.state.json   — PID e porta do simulador em execução
└── jdk/                   — JDK/JRE provisionado automaticamente
    └── bin/java
```

## Justificativa

- Convenção de `~/.nome-ferramenta/` (similar a `~/.aws`, `~/.kube`, `~/.docker`).
- Escopo por usuário: não requer privilégios de administrador.
- Cache de artefatos: evita re-download desnecessário.
- Formato JSON para state files: legível, debugável via `cat`.

## Alternativas consideradas

- **Banco de dados SQLite**: overkill para poucos registros; viola princípio da simplicidade.
- **Variáveis de ambiente**: não persiste entre sessões; inadequado para PID de processo.
- **Arquivo no diretório corrente**: problemático quando o CLI é invocado de diretórios diferentes.

## Consequências

- Primeiro uso cria o diretório automaticamente.
- O usuário pode inspecionar o estado com `cat ~/.hubsaude/*.state.json`.
- JDK é re-usado entre assinador e simulador.
- `~/.hubsaude/` ignorado pelo git (via `.gitignore`).
