# Inventário parcial do ambiente Windows de desenvolvimento — 2026-10-11

## Escopo e limite

Este registro descreve apenas o computador usado nesta sessão de desenvolvimento. Ele não identifica nem homologa o servidor de produção da Majucau. A direção aprovada para compartilhar dados continua sendo hospedagem privada centralizada; esta medição não escolhe provedor, região ou máquina de produção.

## Observações verificadas

Leitura somente, feita pelo Windows PowerShell com consultas CIM e de serviços:

| Item | Resultado na captura |
|---|---|
| Sistema | Windows 10 Pro, versão 10.0.19045, 64 bits |
| Processador | 8 núcleos e 16 processadores lógicos |
| Memória | 15,9 GiB instalados visíveis; 4,1 GiB livres no momento da consulta |
| Unidade C: | 476,4 GiB de capacidade; 90,7 GiB livres no momento da consulta |
| Rede | Windows reportou conectividade IPv4 e IPv6 com a Internet; medição pontual |
| PostgreSQL | Serviço `postgresql-x64-18` encontrado em execução e configurado para iniciar automaticamente |
| Worker Majucau | Nenhum serviço com nome ou descrição Majucau foi encontrado na consulta |

Os valores livres de memória e disco variam com o uso. O nome do computador, contas, endereço de rede, números de série, dados de banco e credenciais foram omitidos. Nenhum serviço, porta, arquivo ou configuração foi alterado.

## Implicações operacionais já aprovadas

- O serviço PostgreSQL encontrado não foi considerado exclusivo da Majucau. Não o reutilizar nem reconfigurar; manter instância separada conforme a decisão registrada.
- A instalação local aprovada mantém banco, worker e aplicativo isolados, com serviço sob identidade própria, inicialização controlada e preflight. Este inventário não substitui validação em instalação limpa.
- Para acesso compartilhado, o alvo continua sendo serviço central autenticado com PostgreSQL privado; os clientes não recebem acesso direto ao banco.

## Situação de suporte do Windows

- A Microsoft informa que o suporte regular do Windows 10 Home/Pro 22H2 terminou em 14/10/2025. Portanto, a versão observada neste computador não deve ser presumida como coberta por atualizações regulares de segurança em 2026.
- O programa ESU pode fornecer atualizações críticas e importantes temporárias, mas exige inscrição. O pacote preparatório KB5126256 sozinho não inscreve o dispositivo nem fornece atualizações. A inscrição deste computador não foi verificada.
- Antes de aprovar uma máquina Windows 10 como host operacional, comprovar a inscrição ESU vigente ou atualizar para uma versão Windows ainda suportada. Este computador continua classificado apenas como ambiente de desenvolvimento, não como host de produção.
- Fontes oficiais consultadas em 11/10/2026: [ciclo de vida do Windows 10](https://learn.microsoft.com/pt-br/lifecycle/announcements/windows-10-22h2-end-of-support-update) e [pacote de preparação ESU KB5126256](https://support.microsoft.com/en-us/servicing/os/windows-10/2026/09/kb5126256-windows-10-21h2-22h2-standalone-cbs).

## O que falta para concluir BK-014

- Confirmar qual equipamento, se houver, será o host local da operação; esta máquina de desenvolvimento não pode ser presumida como host da empresa.
- Verificar energia/continuidade, estabilidade e qualidade da rede, permissões de instalação, rotina de suporte e eventual acesso externo autorizado.
- Fechar a proposta operacional de atualização, backup, retenção, recuperação e monitoramento para o ambiente escolhido e validá-la com a Majucau.
- Para o host escolhido, comprovar inscrição ESU vigente ou escolher uma versão Windows ainda suportada; a inscrição deste computador não foi verificada.

**Progresso registrado:** BK-014 passa de 0% para 15%, pelo inventário parcial e pela verificação oficial do ciclo de suporte do ambiente atual. O host operacional ainda não foi escolhido e o critério de aceite não está concluído.
