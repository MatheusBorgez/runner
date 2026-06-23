# Sistema Runner — Implementação

Implementação do [Sistema Runner](../especificacao.md) para a disciplina de Implementação e Integração (2026-01), UFG/SES-GO.

## Estrutura

```
projetos/
├── go.mod                    ← módulo Go: github.com/kyriosdata/runner
├── cmd/
│   ├── assinatura/           ← CLI assinatura (binário principal)
│   │   ├── main.go
│   │   └── cli/              ← subcomandos Cobra
│   └── simulador/            ← CLI simulador
│       ├── main.go
│       └── cli/              ← subcomandos Cobra
├── internal/
│   ├── invoker/              ← invocação do assinador.jar (local e HTTP)
│   ├── jdk/                  ← provisionamento automático do JDK 21
│   ├── release/              ← download de artefatos via release.json
│   └── runtime/              ← estado persistente em ~/.hubsaude/
└── assinador-java/           ← assinador.jar (Java 21, Maven)
    ├── pom.xml
    └── src/
```

## Como compilar

### CLIs Go

Requer Go 1.26+.

```bash
cd projetos

# Compilar assinatura
go build -o assinatura ./cmd/assinatura

# Compilar simulador
go build -o simulador ./cmd/simulador
```

### assinador.jar (Java)

Requer JDK 21+ e Maven 3.9+.

```bash
cd projetos/assinador-java
mvn clean package
# JAR gerado em: target/assinador-1.0.0-SNAPSHOT.jar
```

## Como executar os testes

### Testes Go

```bash
cd projetos
go test ./...
```

### Testes Java

```bash
cd projetos/assinador-java
mvn test
```

## Como usar

### Copiar o assinador.jar para ~/.hubsaude/

```bash
cp projetos/assinador-java/target/assinador-*.jar ~/.hubsaude/assinador.jar
```

### Fluxo com modo servidor (recomendado)

```bash
# 1. Inicia o servidor
assinatura start

# 2. Cria assinatura (usa servidor por padrão)
assinatura sign --content $(echo -n "meu conteudo" | base64)

# 3. Valida assinatura
assinatura validate --content $(echo -n "meu conteudo" | base64) --signature "MOCKED_SIGNATURE_BASE64_=="

# 4. Verifica status
assinatura status

# 5. Encerra servidor
assinatura stop
```

### Fluxo com modo local (subprocess direto)

```bash
assinatura sign --content $(echo -n "conteudo" | base64) --local
```

### Gerenciar o Simulador HubSaúde

```bash
simulador start         # baixa e inicia simulador.jar
simulador status        # exibe estado
simulador stop          # encerra
```

## Verificar artefatos de release

```bash
cosign verify-blob \
  --certificate assinatura-v1.0.0-linux-amd64.pem \
  --signature  assinatura-v1.0.0-linux-amd64.sig \
  assinatura-v1.0.0-linux-amd64
```

## Rastreabilidade

| Épico | Histórias | Status |
|-------|-----------|--------|
| US-01 | US-01.1 a US-01.9 | Sprint 1 (base) + Sprint 2/3 (sign/validate/servidor) |
| US-02 | US-02.1 a US-02.5 | Sprint 2 (domínio) + Sprint 3 (HTTP + PKCS#11) |
| US-03 | US-03.1 a US-03.4 | Sprint 4 (simulador CLI) |
| US-04 | US-04.1 | Sprint 2 (provisionamento JDK) |
| US-05 | US-05.1 a US-05.3 | Sprint 1 (CI/CD + Cosign) |
