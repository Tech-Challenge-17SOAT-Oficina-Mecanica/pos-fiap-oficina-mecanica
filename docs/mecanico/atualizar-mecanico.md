---
documento: Refinamento de Requisitos — Contexto de Mecânico
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-25
status: rascunho
---

# Refinamento de Requisitos — Atualizar Mecânico

Este documento define a atualização cadastral do profissional e de suas permissões, em integração
transacional com a conta mantida pelo contexto de Segurança.

## 1 · Atualizar Mecânico

### 1.1 Refinamento de Produto

**Persona**

Mecânico autorizado.

**Objetivo**

Atualizar os dados cadastrais e as permissões de um mecânico existente.

**Problema**

A oficina precisa manter os dados dos profissionais atualizados e controlar quais operações cada
mecânico pode executar, preservando a rastreabilidade e evitando permissões indevidas.

**Pré-condições**

- O solicitante está autenticado e possui `mecanicos:escrever`.
- O mecânico existe.
- O solicitante possui a versão atual do cadastro.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-MEC-04 | Permitir atualizar nome, e-mail e escopos de um mecânico. |
| RF-MEC-05 | Impedir que o e-mail seja usado por outro usuário. |
| RF-MEC-06 | Substituir os escopos do usuário pelos escopos informados e permitidos. |
| RF-MEC-07 | Impedir a sobrescrita de atualização concorrente. |
| RF-MEC-08 | Retornar o cadastro atualizado sem expor a senha. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-MEC-03 | A atualização deve ocorrer em uma única transação. |
| RNF-MEC-04 | A operação deve usar controle otimista por versão. |
| RNF-MEC-05 | A senha não deve ser alterada nem retornada nesta operação. |

**Fluxo Principal**

1. O mecânico autorizado informa o identificador, a versão, nome, e-mail e escopos.
2. O sistema valida autenticação, autorização e os dados recebidos.
3. O sistema localiza o mecânico e confirma a versão atual.
4. O sistema verifica a disponibilidade do e-mail e a validade dos escopos.
5. O sistema atualiza usuário, mecânico e escopos na mesma transação.
6. O sistema incrementa a versão e retorna os dados atualizados.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Mecânico inexistente | Informa que o cadastro não foi encontrado. |
| A2 | E-mail pertence a outro usuário | Impede a atualização. |
| A3 | Escopo inválido | Impede a atualização. |
| A4 | `If-Match` ausente | Exige a versão antes de atualizar. |
| A5 | Versão desatualizada | Impede a sobrescrita concorrente. |
| A6 | Usuário sem autorização | Impede a operação. |

**Saída**

- Mecânico atualizado, com dados cadastrais e escopos vigentes.

**Pós-condições**

- Nome, e-mail e escopos refletem a última atualização confirmada.
- A senha permanece inalterada e não é exposta.
- O vínculo do mecânico com Ordens de Serviço permanece preservado.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
PUT /mecanicos/{mecanicoId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Escopo: `mecanicos:escrever`.

**Entrada**

| Origem | Campo | Tipo | Descrição |
|---|---|---|---|
| Path | `mecanicoId` | uuid | Obrigatório; identificador do mecânico. |
| Header | `If-Match` | integer | Obrigatório; versão atual do mecânico. |
| Body | `nome` | string | Obrigatório; nome atualizado. |
| Body | `email` | string | Obrigatório; e-mail de login atualizado. |
| Body | `escopos` | array de string | Obrigatório; escopos substitutos. |

```json
{
  "nome": "Maria Souza",
  "email": "maria.souza@oficina.local",
  "escopos": ["veiculos:ler", "veiculos:escrever", "os:ler"]
}
```

**Validações**

- `mecanicoId` deve ser UUID válido.
- `If-Match` é obrigatório e deve corresponder à versão atual.
- `nome`, `email` e `escopos` são obrigatórios.
- O e-mail deve ter formato válido e não pode pertencer a outro usuário.
- Cada escopo deve pertencer à lista oficial do projeto.

**Processamento**

1. Validar JWT e o escopo `mecanicos:escrever`.
2. Validar identificador, `If-Match` e corpo da requisição.
3. Consultar o mecânico e validar a versão.
4. Validar o e-mail e os escopos informados.
5. Abrir transação.
6. Atualizar e-mail em `usuario`, nome e versão em `mecanico` e os vínculos em `usuario_escopo`.
7. Confirmar a transação e retornar o cadastro atualizado.

**Persistência**

- Consulta: `mecanico`, `usuario` e escopos permitidos.
- Altera: `usuario`, `mecanico` e `usuario_escopo`.
- A versão é incrementada somente quando todas as alterações são confirmadas.
- A senha não é consultada para resposta nem alterada.

**Saída da API**

```json
{
  "id": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
  "nome": "Maria Souza",
  "email": "maria.souza@oficina.local",
  "ativo": true,
  "escopos": ["veiculos:ler", "veiculos:escrever", "os:ler"],
  "version": 2
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Mecânico atualizado. |
| `400` | Identificador, corpo, e-mail ou escopo inválido. |
| `401` | Token ausente, inválido ou expirado. |
| `403` | Token sem `mecanicos:escrever`. |
| `404` | Mecânico não encontrado. |
| `409` | E-mail já cadastrado para outro usuário. |
| `412` | Versão informada não corresponde à versão atual. |
| `428` | Header `If-Match` ausente. |

**Dependências**

- Middleware HTTP de autenticação e autorização.
- `MecanicoRepository`.
- `UsuarioRepository` e serviço de escopos do contexto de Segurança.
- Controle transacional de persistência.

**Testes**

*Unitários*

- Atualiza nome, e-mail e escopos com versão válida.
- Rejeita mecânico inexistente, e-mail duplicado e escopo inválido.
- Rejeita ausência de `If-Match` e versão desatualizada.
- Garante que a senha não seja alterada nem retornada.

*Integração*

- Retorna `200` e persiste usuário, mecânico e escopos na mesma transação.
- Retorna `400`, `401`, `403`, `404`, `409`, `412` e `428` para os cenários mapeados.
- Preserva os vínculos existentes do mecânico com Ordens de Serviço.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Adicionar `version` ao agregado Mecânico e validar a transição de versão.

**Caso de uso**

- [ ] Implementar `AtualizarMecanico` com atualização transacional de cadastro e escopos.

**Repositório**

- [ ] Buscar mecânico por identificador e atualizar nome com controle de versão.
- [ ] Atualizar e-mail do usuário e substituir os escopos vinculados.

**Integrações**

- [ ] Integrar com Segurança para validar escopos permitidos e manter a conta vinculada.

**Handler HTTP**

- [ ] Implementar `PUT /mecanicos/{mecanicoId}` e leitura de `If-Match`.

**Validações**

- [ ] Validar UUID, campos obrigatórios, formato e unicidade do e-mail e escopos permitidos.

**Concorrência**

- [ ] Comparar `If-Match` com `mecanico.version`, retornar `412` em divergência e `428` quando ausente.

**Testes unitários**

- [ ] Cobrir sucesso, inexistência, e-mail duplicado, escopo inválido, ausência de versão e conflito de versão.

**Testes de integração**

- [ ] Cobrir persistência transacional, atualização de escopos, preservação de senha e dos vínculos com OS.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Code Review aprovado.
