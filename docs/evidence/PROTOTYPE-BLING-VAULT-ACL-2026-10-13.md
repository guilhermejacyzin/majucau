# Cofre seguro do protótipo e diagnóstico da conexão Bling — 13/10/2026

## Causa identificada

- A primeira tentativa de salvar a configuração retornou `BLING_VAULT_UNAVAILABLE`.
- O worker informou estado `OK` e PostgreSQL `OK`; o problema era a pasta pai do cofre. A pasta tinha somente uma permissão herdada para o usuário atual, enquanto o DPAPI exige uma lista de acesso protegida e restrita.
- A interface exibia pontos no campo Client Secret por causa de um texto de exemplo. Isso não significava que o segredo estivesse salvo; a validação posterior confirmou que não havia Client Secret disponível no cofre.

## Correção e comprovação

- `scripts/start-prototype.ps1` agora protege a pasta de credenciais antes de iniciar o worker. Usa o SID do usuário local e os SIDs de SYSTEM e Administradores, desativa permissões herdadas e confere que existem exatamente três permissões explícitas de controle total, com herança para arquivos e subpastas.
- Depois de reiniciar o protótipo, a pasta de credenciais e o cofre apresentaram DACL protegida com os três acessos previstos. O worker informou PostgreSQL `OK`.
- O salvamento avançou de `BLING_VAULT_UNAVAILABLE` para `BLING_SECRET_REQUIRED`. Isso confirma que banco e cofre estão acessíveis e que falta cadastrar o Client Secret do aplicativo Bling.

## Limites e próximo passo

- Os dados públicos do aplicativo foram preenchidos apenas no formulário local. O salvamento não foi concluído, porque o Client Secret está ausente; por isso, esses dados ainda não foram persistidos.
- Não consultei nem copiei segredo, senha ou token. Não iniciei OAuth, teste de conexão ou sincronização com a API.
- Para continuar, Gisele deve inserir o Client Secret diretamente no campo do Majucau e salvar. A senha usada para entrar no site do Bling é diferente do Client Secret da integração.
- Esta correção cobre somente o protótipo de desenvolvimento, que executa com a identidade do usuário local. Não comprova o instalador de produção: o NSIS ainda precisa registrar o serviço, provisionar a pasta sob a identidade definitiva e fornecer a credencial do banco sem texto aberto.
- BK-024 permanece em 35%; BK-041 permanece em 75% (25% restante). A média do backlog permanece 10,9% concluída e 89,1% restante. Não há prazo confiável para conexão e sincronização antes de cadastrar o segredo e homologar as páginas reais.

Nenhuma tela aprovada ou regra financeira foi alterada.
