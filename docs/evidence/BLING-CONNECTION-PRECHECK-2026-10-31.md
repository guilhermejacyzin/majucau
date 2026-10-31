# Pré-checagem da conexão Bling no protótipo — 2026-10-31

## Objetivo

Encontrar por que a sincronização real não inicia e abrir o protótipo para configurar a conexão com segurança, sem ler ou registrar segredos.

## Evidências verificadas

- O estado do Git antes da tentativa estava limpo no branch `codex/majucau-artifact-consolidation`, sincronizado com `origin` no commit `3414a77`.
- O script oficial de desenvolvimento `scripts/start-prototype.ps1` usa PostgreSQL 18 isolado em `127.0.0.1:55438`, protege segredos de desenvolvimento via DPAPI e concentra temporários em `artifacts/cache/` e `artifacts/retained/postgresql-dev/`.
- Antes do início, o banco isolado estava parado; os arquivos da instância e a credencial local de banco protegida estavam presentes.
- O cofre efetivamente usado pelo worker de desenvolvimento é `artifacts/retained/postgresql-dev/credentials/worker-vault/`. O arquivo esperado para credenciais Bling (`bling_credentials_v1.dpapi`) não estava presente. O conteúdo de nenhum arquivo protegido foi aberto.
- A execução de `scripts/start-prototype.ps1` compilou o worker, iniciou o banco isolado e o worker, abriu o servidor Vite e iniciou a janela Wails. A árvore acessível inicial mostrou o painel executivo sem valores confirmados. Isso não comprova OAuth, conexão ou sincronização real.
- Não houve chamada OAuth, teste de conexão, consulta à API Bling ou leitura de dados financeiros. A leitura da página `www.bling.com.br` foi recusada pela política de acesso do navegador; nenhuma alternativa foi tentada para contornar essa recusa.
- A navegação automática para Integrações não foi possível: o clique acessível não tinha geometria de coordenadas e duas capturas subsequentes da janela expiraram. A janela permanece aberta no painel para a responsável navegar manualmente.

## Próxima ação necessária

Na janela Majucau aberta, a responsável deve abrir `Integrações`, informar o **Client Secret do aplicativo Bling** no campo próprio e salvar a configuração. Não enviar o segredo pelo chat. A senha usada para entrar no site Bling não é o Client Secret.

Após essa gravação, verificar somente o estado configurado sem revelar o conteúdo do cofre, executar `Testar conexão` e então prosseguir com OAuth e uma sincronização de leitura autorizada. Qualquer etapa que abra uma tela de autenticação precisa ser feita pela responsável.

BK-041 está **80% concluída, com 20% restante**. A média do projeto segue em **11,8% concluído, com 88,2% restante**.

## Arquivos temporários e limpeza

- Os logs, o worker e o banco temporário permanecem sob `artifacts/retained/postgresql-dev/` enquanto a janela estiver aberta; caches de build ficam sob `artifacts/cache/`.
- O roteiro oficial encerra o worker e o banco que iniciou, e remove o worker temporário e `frontend/dist/` quando a janela Wails fecha.
- O Wails atualizou `frontend/package.json.md5` durante a verificação de dependências. O repositório estava limpo antes da execução; após encerrar a janela, restaurar esse arquivo ao conteúdo versionado e confirmar o estado do Git. Não remover o banco ou os arquivos protegidos de `artifacts/retained/postgresql-dev/` durante o desenvolvimento.

Nenhuma tela aprovada ou regra financeira foi modificada.
