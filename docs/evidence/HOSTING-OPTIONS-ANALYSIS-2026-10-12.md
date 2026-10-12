# Análise: hospedagem central ou servidor local — 12/10/2026

## Resumo para decisão

Para o Majucau, a direção mais adequada é **manter o protótipo local durante o desenvolvimento e preparar a versão de uso real em hospedagem privada centralizada**. Isso segue a ADR-002 já aprovada e atende melhor ao uso por pessoas em locais diferentes com uma única base compartilhada.

O Google Drive deve continuar sendo a origem dos arquivos externos. Ele não deve ser usado como banco de dados do Majucau. A aplicação e o PostgreSQL devem ter uma infraestrutura própria, com o banco acessível somente pelo serviço do Majucau.

Esta análise apoia a arquitetura existente. Não escolhe provedor, não contrata hospedagem, não altera tela ou regra financeira e não provisiona contas.

## Cenário considerado

- Gisele terá o perfil administrativo central do Majucau.
- Os Diretores precisam consultar e exportar relatórios e backup; a implementação das exportações fica para depois da finalização do trabalho já solicitado.
- Os documentos externos ficam no Google Drive, na pasta compartilhada `Majucau`, com as subpastas `01_perdas_operacionais`, `02_nuvem_pago/recebimentos_futuros` e `03_nuvem_envio`.
- Foi solicitado que a atualização diária ocorra às 09:30, após os arquivos serem colocados nas pastas.
- O protótipo atual usa aplicativo local Wails/Windows e PostgreSQL local dedicado.

## Comparação simples

| Opção | O que ajuda | O que exige | Adequação ao Majucau |
|---|---|---|---|
| **Servidor na empresa** | Os dados ficam num equipamento da empresa e não há cobrança mensal de hospedagem. | Um computador/servidor precisa ficar ligado e ser mantido; acesso de outro local exige uma conexão remota segura, além de cuidar de energia, internet, atualizações, cópias de segurança e recuperação. | Só seria preferível se a empresa já tivesse servidor e suporte técnico contínuo. O cenário informado apresenta o Drive como local compartilhado, não um servidor de aplicação disponível. |
| **Hospedagem privada centralizada** | Uma aplicação e uma base atendem os dois perfis; a rotina das 09:30 pode rodar no serviço central sem depender do computador de Gisele estar ligado; atualizações e backup ficam centralizados. | Há custo recorrente e dependência de internet. É preciso escolher provedor, região, forma segura de acesso, retenção e responsável pela operação. | **Mais adequada**, desde que a aplicação seja protegida e o PostgreSQL não fique exposto à internet. |

## Como o fluxo recomendado funcionaria

1. Os arquivos são colocados nas pastas aprovadas do Google Drive.
2. No horário combinado, o serviço central lê somente os arquivos autorizados, valida cada registro e informa se a atualização terminou ou encontrou problema.
3. O serviço grava os dados na base central usando as regras já aprovadas de streaming, atomicidade e idempotência.
4. Gisele entra com o perfil administrativo; Diretores consultam e, após a etapa combinada, exportam os relatórios/backup conforme as permissões aprovadas.
5. Os computadores dos usuários acessam somente a aplicação autenticada. Eles não recebem senha do PostgreSQL nem conexão direta ao banco.

A documentação oficial do Drive descreve a obtenção do conteúdo de arquivos pela API e recomenda verificar se o usuário pode baixá-los. Isso permite tratar o Drive como fonte de arquivos. [Google Drive API — baixar e exportar arquivos](https://developers.google.com/workspace/drive/api/guides/manage-downloads)

Como exemplo de configuração de provedor, o Google Cloud documenta PostgreSQL com IP privado e os requisitos de rede para que o serviço autorizado se conecte. Isso é apenas um exemplo técnico: **não é uma escolha do Google Cloud para o projeto**. [Google Cloud SQL — configurar IP privado](https://docs.cloud.google.com/sql/docs/postgres/configure-private-ip?hl=en)

O princípio de menor privilégio recomenda dar a cada usuário somente as permissões necessárias e validar cada acesso. Para este sistema, a aplicação deve conferir o perfil no serviço; esconder um botão na tela não substitui essa verificação. [Google Cloud — princípio de confiança zero](https://docs.cloud.google.com/architecture/framework/security/implement-zero-trust?hl=en)

## O que ainda precisa ser decidido antes da hospedagem real

- Provedor, região e orçamento mensal.
- Se o acesso dos usuários será pela internet com autenticação forte ou por rede privada/VPN.
- Fuso horário do agendamento das 09:30 e o que fazer quando algum arquivo estiver atrasado ou ausente; o serviço deve preservar a última informação válida enquanto isso não estiver definido.
- Autorização de leitura do Drive limitada às pastas necessárias, além da forma de guardar e renovar a credencial.
- MFA, sessão e recuperação de acesso.
- Retenção de backups cifrados e comprovação de restauração.
- Aplicação efetiva das permissões no serviço e validação a partir dos dois locais.

O custo e o prazo não podem ser estimados com responsabilidade sem provedor, região, volume de dados, retenção e disponibilidade desejada.

## Observação sobre permissões dos Diretores

A confirmação mais recente recebida para esta análise é **consulta e exportação**, com o trabalho de exportação adiado até a finalização do pedido atual. O registro anterior também permitia aos Diretores sincronizar e importar arquivos; essa diferença precisa ser reconciliada no registro oficial de permissões antes de qualquer alteração de autorização no produto. Esta análise não muda as permissões já implementadas nem libera a exportação.

## Estado

Recomendação: manter o protótipo local enquanto ele é homologado; depois, seguir a hospedagem privada centralizada aprovada na `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md`. Ainda não é recomendação de provedor nem autorização para publicar dados reais.
