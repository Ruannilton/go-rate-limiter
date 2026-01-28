#  Objetivo Principal
Montar um rate limiter em golang que permita ser altamente configurável, isto é permitir alterar qual algoritmo será utilizado, a forma de armazenamento e a quantidade de acessos que um usuário pode fazer a um determinado recurso.

## Objetivos Secundários
- Montar um painel de configuração
- Montar uma web api para testes
- Adicionar um painel de métricas

### Algoritmos Suportados
- Token Bucket
- Leaky Bucket
- Fixed Window Counter
- Fixed Window Log

### Armazenamentos Suportados
- Memória
- Persistido Local
- Distribuido

## Regras Funcionais
- Cadastrar uma regra (chave do recurso, algoritmo, armazenamento, quantidade de acessos)
- Atualizar uma regra (chave do recurso, algoritmo, armazenamento, quantidade de acessos)
- Excluir uma regra (chave do recurso)
- Listar as regras
- Executar uma regra (identificador, chave do recurso)
