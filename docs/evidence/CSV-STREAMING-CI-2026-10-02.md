# Importação CSV em streaming — evidência de CI

- Commit: `d66954d5d2135586fcd42dc52363e1b3640a1fce`
- GitHub Actions run: [37090169223](https://github.com/guilhermejacyzin/majucau/actions/runs/37090169223)
- Resultado: `verify-windows` e `verify-postgres` concluídos com sucesso.
- Windows: Go tests, `go vet`, varredura de vulnerabilidades, checksums de migrations, geração SQL, scan de secrets, lint/typecheck/testes/build do frontend, auditoria de dependências, builds desktop/worker e smoke de instalação/desinstalação passaram.
- PostgreSQL: migrations foram aplicadas em ordem num banco descartável e os testes de integração passaram.
- Escopo validado: importações de produção Bling/Nuvem Pago validam e persistem registros um a um na transação; IDs externos repetidos são deduplicados por HMAC temporário; linhas inválidas resultam em `PARTIAL`, e erro técnico, cancelamento, CSV estruturalmente ambíguo ou limite excedido revertem o lote.
- Limites aprovados: 1 MiB por campo, 4 MiB por registro e 256 campos; sem limite total de linhas. A fila de nomes de arquivos lê no máximo 256 nomes por bloco.
- Limitações: esta evidência não conclui DATA-04. Pré-visualizações continuam materializando dados/erros em memória; backup ainda lê dumps/payloads completos; falta auditar outros fluxos e verificar ACL/retensão dos temporários numa instalação real.
- Artefato: `majucau-g1-unsigned`, ID `11262108080`, 36.9 MB. O artefato remoto não pode ser apagado pelas ferramentas conectadas e está listado no registro local `artifacts/retained/cleanup-register.md` para exclusão manual.
