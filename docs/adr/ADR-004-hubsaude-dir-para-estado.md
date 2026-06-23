# ADR-004 — Diretório ~/.hubsaude/ para estado e artefatos

**Status:** Aceito | **Contexto:** US-01.5, US-03, US-04

## Decisão

Usar `~/.hubsaude/` como diretório base para estado, JDK provisionado e JARs.

```
~/.hubsaude/
├── assinador.jar
├── simulador.jar
├── simulador.version
├── assinador.state.json   ← PID e porta do assinador em execução
├── simulador.state.json   ← PID e porta do simulador em execução
└── jdk/
    └── bin/java
```

## Justificativa

- Convenção `~/.<ferramenta>/` (cf. `~/.aws`, `~/.kube`).
- Escopo por usuário: sem privilégios de administrador.
- Formato JSON: legível com `cat`, fácil de debugar.

## Alternativas rejeitadas

- **SQLite**: overkill para poucos registros.
- **Variáveis de ambiente**: não persistem entre sessões.
- **Diretório corrente**: quebra quando o CLI é invocado de locais diferentes.
