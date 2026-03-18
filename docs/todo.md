# Sistema Runner — Etapas de Desenvolvimento (TODO)

Este documento consolida todas as etapas de desenvolvimento do **Sistema Runner** com base na especificação, no design e no planejamento do repositório. Serve como lista de tarefas (to-do) para guiar a construção de forma iterativa e incremental.

---

## Resumo do problema e da solução

**Problema:** Facilitar o acesso à execução de aplicações Java via linha de comandos, sem que o usuário precise conhecer detalhes de configuração ou instalação do ambiente Java. O trabalho é de interesse da SES-GO e UFG no contexto de interoperabilidade de dados em saúde.

**Solução:** Dois CLIs multiplataforma (**assinatura** e **simulador**) que orquestram duas aplicações Java (**assinador.jar** e **simulador.jar**). O assinador simula criação e validação de assinatura digital com validação rigorosa de parâmetros (FHIR); o simulador é obtido via GitHub Releases e seu ciclo de vida é gerenciado pelo CLI. JDK pode ser provisionado automaticamente quando ausente. Binários são distribuídos via GitHub Releases com versionamento semântico e assinatura Cosign/Sigstore.

---

## 0. Preparação e ambiente

### Stack e build (definido)

- [x] **Linguagem e build:** Java 17 + Maven para todo o sistema (**assinatura**, **simulador** e **assinador.jar**). Stack única facilita manutenção e reuso de código (ex.: validações, HTTP).
- [ ] Configurar **estrutura Maven** no padrão Java/Maven (layout de diretórios `src/main/java`, `src/test/java`, `pom.xml`). Estrutura atual: **multi-módulo** com `runner-parent` e módulos em `implementacao/` (`assinador`, `assinatura`, `simulador`) para build e versão coordenados.
- [ ] Definir **convenções de código**: estilo (ex.: Google Java Style ou padrão Maven), **Checkstyle** e **SpotBugs** (ou similar) no build; quebrar build em caso de violação (opcional mas recomendado).

### Fluxo Git (definido)

- [x] **Branches:**
  - **`main`** — produção (código estável, pronto para release).
  - **`develop`** — homologação / próximo release; integração de novas funcionalidades.
  - **`feat/descricao-feat`** — nova funcionalidade (ex.: `feat/assinatura-criar`, `feat/provisionar-jdk`). Merge em `develop`.
  - **`fix/descricao-fix`** — correção em funcionalidade já existente (ex.: `fix/validacao-fhir`). Merge em `develop` (ou em `main` se for hotfix crítico; aí definir política de cherry-pick para `develop`).
- [ ] **Proteção de branches:** em GitHub/GitLab, proteger `main` e `develop` (ex.: merge apenas via PR, exigir status de CI verde). Definir se `main` recebe merge só de `develop` ou também de hotfix.
- [ ] **Revisão:** PRs revisados antes do merge; documentar isso em CONTRIBUTING ou no README.

### Ambiente e ferramentas

- [ ] **Ambiente de build:** JDK 17 instalado (ou via SDKMAN/actions no CI); Maven 3.8+; garantir que `mvn clean install` rode testes e relatórios (Checkstyle/SpotBugs).
- [ ] **Diagramas C4:** garantir que `geraimagens.sh` / `geraimagens.bat` gerem as imagens a partir dos `.puml` (útil para documentação e entregáveis).

---

## 1. Aplicação assinador.jar (Java)

### 1.1 Estrutura e contrato

- [ ] Criar projeto Java (Maven/Gradle) para **assinador.jar**
- [ ] Definir contrato de entrada/saída para **criação de assinatura** (ex.: parâmetros FHIR, formato de resposta)
- [ ] Definir contrato de entrada/saída para **validação de assinatura**
- [ ] Definir modo de execução: CLI (invocação direta) e servidor HTTP (API para o CLI chamar)

### 1.2 Validação de parâmetros (US-02)

- [ ] Implementar validação de parâmetros conforme especificações FHIR para criação de assinatura
- [ ] Implementar validação de parâmetros para validação de assinatura
- [ ] Retornar mensagens de erro claras e estruturadas para parâmetros inválidos
- [ ] Tratamento de exceções e propagação estruturada de erros

### 1.3 Simulação de operações (US-02)

- [ ] Simular **criação de assinatura**: resposta pré-construída quando parâmetros válidos
- [ ] Simular **validação de assinatura**: resultado pré-determinado (válido/inválido) com critérios simples
- [ ] Preparar interface/estrutura para suporte a **PKCS#11** (dispositivo criptográfico), mesmo que não implementado de forma real

### 1.4 Modo servidor HTTP

- [ ] Expor endpoints HTTP para criação e validação de assinatura (mesmo fluxo lógico do modo CLI)
- [ ] Documentar contrato da API (ex.: paths, métodos, payloads) para integração com o CLI

### 1.5 Testes (assinador.jar)

- [ ] Testes unitários para validadores de parâmetros
- [ ] Testes unitários para simulação de criação/validação
- [ ] Testes de integração (modo CLI e modo HTTP)
- [ ] Casos de teste para cenários de erro (parâmetros inválidos, exceções)

---

## 2. Aplicação assinatura (CLI)

### 2.1 Núcleo e invocação do Assinador (US-01)

- [ ] Implementar CLI que aceite comandos para **criação** e **validação** de assinatura
- [ ] Implementar **invocação direta** do assinador.jar (ex.: `java -jar assinador.jar ...`)
- [ ] Implementar **invocação via HTTP**: enviar requisições ao Assinador em modo servidor
- [ ] Validar entrada do usuário no CLI antes de chamar o Assinador
- [ ] Formatar e exibir resultado da operação de forma legível (sucesso e erro)

### 2.2 Tratamento de erros (especificação §6.3)

- [ ] Capturar erros em qualquer ponto do fluxo (CLI e resposta do Assinador)
- [ ] Propagar erros de forma estruturada
- [ ] Apresentar ao usuário mensagens claras com informação suficiente para correção

### 2.3 Provisionamento de JDK (US-04)

- [ ] Detectar se o JDK está presente na máquina na versão exigida
- [ ] Baixar JDK compatível quando ausente (Windows, Linux, macOS)
- [ ] Disponibilizar o JDK baixado para uso pelo Assinador (e depois pelo Simulador)
- [ ] Testar download nas três plataformas

### 2.4 Testes (CLI assinatura)

- [ ] Testes unitários para parsing de argumentos e validação de entrada
- [ ] Testes de integração: CLI → assinador.jar (modo direto e modo HTTP)
- [ ] Testes de aceitação baseados nos critérios de US-01 e US-02

---

## 3. Aplicação simulador (CLI)

### 3.1 Obtenção do simulador.jar (US-03)

- [ ] Implementar download do **simulador.jar** a partir do repositório da disciplina (GitHub Releases)
- [ ] Baixar apenas a versão mais recente; não baixar de novo se já existir localmente a mesma versão
- [ ] Tratar falhas de rede e releases inexistentes com mensagens claras

### 3.2 Ciclo de vida do Simulador (US-03)

- [ ] Comando para **iniciar** o Simulador (usando JDK provisionado ou do sistema)
- [ ] Verificar se as portas necessárias estão disponíveis antes de iniciar
- [ ] Comando para **parar** o Simulador
- [ ] Comando ou saída para **exibir status** do Simulador (em execução ou não)
- [ ] Tratamento de erros (porta em uso, processo não encontrado, etc.)

### 3.3 Reuso de JDK (US-04)

- [ ] Reutilizar o mesmo mecanismo de provisionamento de JDK do CLI assinatura (ou biblioteca compartilhada) para executar simulador.jar

### 3.4 Testes (CLI simulador)

- [ ] Testes unitários para lógica de download e verificação de versão local
- [ ] Testes de integração para start/stop/status (com simulador.jar real ou mock)
- [ ] Testes de aceitação para US-03

---

## 4. Binários multiplataforma e distribuição (US-05)

- [ ] Definir versionamento semântico (SemVer) e número de versão inicial (ex.: 1.0.0)
- [ ] Build de **assinatura** para Windows (amd64), Linux (amd64), macOS (amd64)
  - [ ] `assinatura-<ver>-windows-amd64.exe`
  - [ ] `assinatura-<ver>-linux-amd64.AppImage`
  - [ ] `assinatura-<ver>-macos-amd64.dmg`
- [ ] Build de **simulador** para as mesmas plataformas
  - [ ] `simulador-<ver>-windows-amd64.exe`
  - [ ] `simulador-<ver>-linux-amd64.AppImage`
  - [ ] `simulador-<ver>-macos-amd64.dmg`
- [ ] Gerar checksums SHA256 para todos os artefatos
- [ ] Publicar releases no **GitHub Releases** com artefatos e checksums

---

## 5. Integridade e assinatura de artefatos (Cosign/Sigstore)

- [ ] Configurar **Cosign** para assinatura com identidade OIDC e transparency log Sigstore
- [ ] Para cada artefato na release, publicar: `<artefato>`, `<artefato>.sig`, `<artefato>.pem`
- [ ] Integrar assinatura no **pipeline de CI/CD** (automático na criação da release)
- [ ] Documentar para o usuário como verificar artefatos com `cosign verify-blob`

---

## 6. Código fonte do Simulador do HubSaúde (entregável 7)

- [ ] Confirmar com o professor o escopo: a especificação menciona que simulador.jar “não faz parte do escopo de desenvolvimento”, mas os entregáveis incluem “Código fonte do Simulador”
- [ ] Se aplicável: implementar **simulador.jar** (aplicação Web gerenciada pelo CLI), documentada e compatível com Windows, Linux e macOS

---

## 7. Documentação

- [ ] **Manual do usuário** para o CLI **assinatura** (comandos, exemplos, modos direto/HTTP)
- [ ] **Manual do usuário** para o CLI **simulador** (iniciar, parar, status)
- [ ] **Documentação técnica da integração** (assinatura ↔ assinador.jar; simulador ↔ simulador.jar)
- [ ] **Exemplos de uso** (scripts ou passos copiáveis)
- [ ] **Guia de instalação** (como obter binários, verificar assinaturas, dependências)
- [ ] Atualizar/corrigir referências no projeto (ex.: `.github/copilot-instructions.md` menciona `contexto.md`; no repositório atual a base é `especificacao.md` e `design.md`)

---

## 8. Especificação e design (entregáveis 5 e diagramas)

- [ ] Manter **especificação** (especificacao.md) alinhada ao que foi implementado (contexto, escopo, requisitos)
- [ ] Completar/refinar requisitos não funcionais se necessário (ex.: desempenho, portabilidade — ISO 25010)
- [ ] Manter **diagramas C4** (contexto e contêineres) atualizados e gerando imagens via scripts existentes
- [ ] Opcional: enriquecer especificação com DoD (Definition of Done) e DoR (Definition of Ready) por user story

---

## 9. CI/CD e qualidade

- [ ] Pipeline de CI: build, testes (unitários e integração), lint
- [ ] Pipeline de release: build multiplataforma, geração de checksums, assinatura Cosign, publicação no GitHub Releases
- [ ] Revisão de código (code review) antes de merge, conforme fluxo Git definido
- [ ] Garantir boa cobertura de testes e tratamento de erros em todo o fluxo

---

## Ordem sugerida (iterações)

1. **Iteração 0:** Preparação (item 0) + estrutura mínima do assinador.jar e do CLI assinatura (itens 1.1, 2.1 parcial).
2. **Iteração 1:** Validação e simulação no assinador.jar (1.2, 1.3) + invocação direta e exibição de resultado no CLI (2.1 completo) + testes (1.5, 2.4 básicos).
3. **Iteração 2:** Modo HTTP do assinador (1.4) + invocação via HTTP no CLI (2.1) + tratamento de erros (2.2) + JDK automático (2.3).
4. **Iteração 3:** CLI simulador (3.1–3.4) + testes.
5. **Iteração 4:** Build multiplataforma (4), Cosign (5), CI/CD (9), documentação (7) e especificação/design (8).
6. **Conforme definido:** Item 6 (código fonte do Simulador do HubSaúde) se for escopo confirmado.

---

*Documento gerado a partir de: especificacao.md, design.md, docs/planejamento.md, README.md, .github/copilot-instructions.md. Atualize este TODO conforme o progresso do projeto.*
