# Limites de chamadas Bling — 2026-10-10

## O que foi implementado

- O adaptador espera pelo menos 400 ms entre inícios de chamadas gerais, mantendo margem abaixo do teto registrado de 3 chamadas por segundo.
- Chamadas ao endpoint `/oauth/token` respeitam pelo menos 3,1 s entre si, atendendo com margem ao limite adicional de 20 chamadas por minuto registrado no mapa de integração.
- A fila protege também chamadas simultâneas: somente uma chamada por vez chega ao transporte dentro de cada sessão do adapter.
- O serviço OAuth de produção cria um limitador compartilhado pelos clientes OAuth e API. A construção de clientes independentes mantém limitadores independentes.
- Cancelamento antes ou durante a espera impede que a requisição chegue ao transporte.
- Falhas de transporte continuam consumindo espaço, pois a chamada já foi iniciada.

## Verificação executada

Com Go 1.26.6 e sem banco de dados configurado:

```text
go test ./internal/integrations/bling -count=1
ok majucau.local/financial-intelligence/internal/integrations/bling 5.536s
```

`go vet ./internal/integrations/bling` também terminou sem apontamentos.

Os testes cobrem o intervalo geral, o limite específico de tokens, a combinação entre chamadas OAuth e API, oito chamadas concorrentes, o cancelamento durante uma espera e o cancelamento enquanto outra requisição ocupa a fila. A última execução da suíte terminou em 5,536 s.

## Limites da evidência

- O limite diário registrado de 120.000 chamadas não é contado pelo adapter. Respostas HTTP 429 continuam sendo tratadas pela camada existente; o checkpoint não deve avançar sem confirmação do lote.
- A evidência é local e automatizada. Não comprova homologação com uma conta Bling real nem substitui confirmação dos limites vigentes na [documentação de limites do Bling](https://developer.bling.com.br/limites).
- A paginação incremental segue sem contrato homologado. Não foi alterada.
- Nenhuma tela ou regra financeira foi alterada.
