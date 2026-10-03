# Limites de registro CSV — evidência de CI

- Commit: `6383d7a9d11c87e0c65f851b1cd5d3b4a90d1cc2`
- GitHub Actions run: [37086256581](https://github.com/guilhermejacyzin/majucau/actions/runs/37086256581)
- Resultado: `verify-windows` e `verify-postgres` concluídos com sucesso.
- Verificações relevantes: testes Go, `go vet`, varredura de vulnerabilidades, checagem de migrations, lint/typecheck/build/testes do frontend, auditoria de dependências, build desktop/worker e smoke do instalador.
- Implementação: o leitor CSV compartilhado aplica limite de 1 MiB por campo, 4 MiB por registro e 256 campos durante a leitura. Os fluxos cobertos são os relatórios CSV de recebimentos do Bling e recebíveis futuros do Nuvem Pago.
- Política aprovada: sem limite total de linhas; processar em blocos/streaming.
- Limitação: esta evidência confirma os limites configurados e a compilação/suíte existentes. Ela não comprova ainda streaming ponta a ponta dos CSVs, pois parsers e importadores mantêm coleções em memória.
- Saída de CI: artefato `11260751393` (`majucau-g1-unsigned`, 36.8 MB); registrado para remoção manual em `artifacts/retained/cleanup-register.md` porque as ferramentas conectadas não oferecem exclusão de artefatos remotos.
