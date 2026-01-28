# GEMINI.md

## Estrutura do Projeto

O projeto está organizado da seguinte forma:

- **cmd/main/**: Contém o ponto de entrada da aplicação (main.go).
- **internal/**: Implementa a lógica principal dos rate limiters, incluindo:
  - `evaluator.go`: Avaliação e controle de limites.
  - `fixed_window.go`: Implementação do algoritmo Fixed Window.
  - `leaky_bucket.go`: Implementação do algoritmo Leaky Bucket.
  - `token_bucket.go`: Implementação do algoritmo Token Bucket.
  - `interfaces.go`: Interfaces comuns para abstração dos limitadores.
- **tests/**: Local destinado aos testes das funcionalidades implementadas.
- **bin/**: Diretório para binários gerados.
- **go.mod**: Gerenciamento de dependências do Go.
- **README.md**: Instruções gerais e informações rápidas do projeto.

## Objetivos

O objetivo deste projeto é fornecer implementações eficientes e reutilizáveis de algoritmos de rate limiting em Go, permitindo o controle de fluxo de requisições em sistemas distribuídos, APIs e aplicações de alta concorrência.

## Funcionalidades

- **Fixed Window Rate Limiter**: Limita o número de requisições em janelas de tempo fixas.
- **Leaky Bucket Rate Limiter**: Controla o fluxo de requisições de forma contínua, simulando um "balde furado".
- **Token Bucket Rate Limiter**: Permite rajadas de requisições até um limite, reabastecendo tokens ao longo do tempo.
- **Interfaces Abstratas**: Facilita a extensão e integração de novos algoritmos de rate limiting.

## Descrição dos Testes

Os testes estão localizados na pasta `tests/` e têm como objetivo:

- Verificar o correto funcionamento de cada algoritmo de rate limiting.
- Garantir que os limites são respeitados em diferentes cenários de carga.
- Testar condições de contorno, como estouro de limites e reinicialização de janelas/buckets.
- Validar a thread safety das implementações.

Sugere-se que os testes incluam:
- Testes unitários para cada algoritmo.
- Testes de concorrência para simular múltiplas requisições simultâneas.
- Testes de performance para avaliar o impacto dos limitadores.

---

Este documento serve como referência para desenvolvedores e revisores, facilitando a compreensão e evolução do projeto.