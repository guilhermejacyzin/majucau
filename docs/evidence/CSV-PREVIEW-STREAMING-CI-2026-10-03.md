# Prévia CSV em streaming — evidência de CI

- Commit validado: `797542ea86fce067859033cd5c08d1332383b22c` (inclui a implementação em `d4a626d`).
- GitHub Actions: [run 37093978791](https://github.com/guilhermejacyzin/majucau/actions/runs/37093978791).
- Resultado: `verify-windows` e `verify-postgres` concluídos com sucesso.
- Windows: testes Go, `go vet`, varredura de vulnerabilidades e secrets, geração SQL, lint/typecheck/testes/build do frontend, builds Wails/worker/helper, contrato e smoke do instalador passaram.
- PostgreSQL: schema aplicado em banco descartável e testes de integração passaram.
- Implementação avaliada: prévias Bling e Nuvem Pago leem as linhas em fluxo sem exigir PostgreSQL; o índice temporário HMAC faz deduplicação exata, e os nomes/localizações temporários são cifrados. A prévia limita metadados e problemas enviados pelo IPC a 50, preserva contagens totais e remove o diretório temporário ao terminar.
- Interface: mudou somente a fonte da contagem total de arquivos; textos, estrutura visual e regras de negócio aprovadas foram preservados.
- Limitações: a suíte atual não possui casos dedicados ao novo `DiskSourceIDSet` e às prévias locais; a aprovação da CI comprova compilação e passagem da suíte existente, não substitui esses casos específicos. Continuam pendentes validar ACL/limpeza em instalação Windows real, converter o backup para streaming e auditar os demais fluxos. DATA-04 permanece `PARTIAL`.
- Artefato remoto: `majucau-g1-unsigned`, ID `11263239567`, 37,073,732 bytes; expira em 2027-01-01 e está registrado em `artifacts/retained/cleanup-register.md` para exclusão manual.
