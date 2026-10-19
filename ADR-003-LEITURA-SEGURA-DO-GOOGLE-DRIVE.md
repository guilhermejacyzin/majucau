# ADR-003 — Leitura segura e somente leitura do Google Drive

- **Status:** aprovado para implementação arquitetural
- **Data:** 2026-10-18
- **Aprovação:** Gisele autorizou escolher e executar a alternativa mais adequada após a análise da arquitetura
- **Escopo:** ingestão das três pastas externas já aprovadas; não altera telas nem regras financeiras

## Contexto

Os arquivos da Majucau ficam no Google Drive e são acessados atualmente pelo Site. O worker local não tem uma pasta sincronizada nem uma autorização persistente. A atualização automática precisa funcionar às 09:30 (`America/Sao_Paulo`) sem depender de uma pessoa manter o navegador aberto. Os arquivos confirmados incluem CSV, TSV, JSON, XLS e XLSX.

As pastas aprovadas são `01_perdas_operacionais`, `02_nuvem_pago/recebimentos_futuros` e `03_nuvem_envio`, dentro de `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`. A estrutura real inclui subpastas autorizadas, como `03_nuvem_envio/Frete`; a leitura percorre somente descendentes encontrados dentro de uma das três raízes. Os IDs reais foram identificados na conta, mas não são registrados no Git. Cada contrato de arquivo ainda precisa de homologação. XLS e XLSX continuam pendentes até existir leitor seguro integrado e os contratos estarem aprovados.

## Decisão

O serviço privado Go do Majucau acessará a API oficial do Google Drive como uma identidade técnica dedicada. Essa identidade receberá papel **Leitor** somente nas três pastas aprovadas. Não terá delegação de domínio, acesso à raiz inteira, nem permissão para criar, editar, mover, renomear ou excluir arquivos.

O Site continua responsável somente pelo fluxo de autorização OAuth do Bling. O navegador não transporta nem interpreta arquivos financeiros. O worker busca os metadados e o conteúdo diretamente da API do Drive, em fluxo, e faz a importação no armazenamento interno do Majucau.

Na hospedagem privada futura, preferir uma identidade de workload/credencial gerenciada pelo provedor escolhido, sem chave privada permanente quando houver suporte. No protótipo local Windows, qualquer credencial necessária deverá ser guardada no cofre protegido por DPAPI e ACL; nunca em código, arquivo de configuração comum, logs, Git ou conversa.

## Controles obrigatórios

1. Configurar os IDs explícitos das três raízes identificadas e limitar a busca a essas raízes; seguir subpastas apenas quando listadas diretamente dentro delas, com paginação e filtro de arquivos na lixeira.
2. Usar apenas operações de listagem, leitura de metadados e download; implementar chamadas externas por caminhos GET revisados e retry idempotente padrão.
3. Conferir metadados antes e depois do download; cancelar o arquivo se sua versão mudar durante a leitura.
4. Calcular SHA-256 durante o fluxo, aplicar limites de tamanho/registro e gravar staging em blocos limitados.
5. Manter atomicidade por lote, replay idempotente, auditoria sanitizada e propagação de cancelamento. Erros de linha deixam lote parcial; falha técnica ou cancelamento reverte o lote.
6. Manter arquivos originais intocados. Fonte ausente permanece pendente; último dado válido não pode ser apresentado como atual.
7. XLS/XLSX não pode produzir fatos até haver parser auditado, limites de conteúdo expandido e dimensões, testes adversos e aprovação dos contratos de aba/cabeçalho/valores. Nenhuma macro será executada.

## Alternativas avaliadas

- **Manter o processamento no Site/Cloudflare:** já existe para CSV/TSV/JSON, mas não resolve com segurança XLS/XLSX e mantém a interpretação de arquivos financeiros na ponte web. Não será a arquitetura de destino.
- **Sincronizar a pasta para o computador:** depende de sessão/configuração local, cria cópias sensíveis fora do fluxo do sistema e não atende bem o serviço central agendado. Rejeitada.
- **Conta técnica com acesso somente de leitura às pastas:** permite execução sem navegador aberto e limita a exposição por identidade e permissões da origem. Escolhida.

## Pendências para a integração real

- Criar/configurar projeto Google Cloud e habilitar a API Drive para o serviço Majucau.
- Criar a identidade técnica e compartilhar **somente** as três pastas aprovadas com papel Leitor; não habilitar delegação de domínio.
- Guardar os IDs das pastas na configuração local protegida e guardar a credencial pelo mecanismo seguro do ambiente.
- Confirmar que a política atual da conta permite compartilhar as pastas com essa identidade.
- Provisionar a credencial técnica, integrá-la ao worker e ao cofre DPAPI, autenticar e validar a leitura real nas três árvores e a retomada do inventário.
- Implementar e homologar o parser seguro XLS/XLSX.
- Homologar amostras sanitizadas e os contratos de cada pasta antes de promover linhas a fatos financeiros.
- Validar permissões, memória, limites, retomada e limpeza em VM limpa.

Se a conta do Drive impedir o compartilhamento restrito com uma identidade técnica, parar antes de ampliar permissões e registrar nova decisão. Não retornar automaticamente à ponte web ou a uma pasta sincronizada.

## Referências técnicas

- [Google Workspace — criar credenciais; pastas específicas podem ser compartilhadas diretamente com a conta de serviço](https://developers.google.com/workspace/guides/create-credentials)
- [Google OAuth — autenticação servidor a servidor e proteção de chaves](https://developers.google.com/identity/protocols/oauth2/service-account)
- [Google Drive API — escopos de acesso](https://developers.google.com/workspace/drive/api/guides/api-specific-auth)
- [Google Drive API — listar arquivos e filtrar por pasta](https://developers.google.com/workspace/drive/api/guides/search-files)
- [Google Drive API — baixar conteúdo](https://developers.google.com/workspace/drive/api/guides/manage-downloads)
