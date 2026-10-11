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

## O que falta para concluir BK-014

- Confirmar qual equipamento, se houver, será o host local da operação; esta máquina de desenvolvimento não pode ser presumida como host da empresa.
- Verificar energia/continuidade, estabilidade e qualidade da rede, permissões de instalação, rotina de suporte e eventual acesso externo autorizado.
- Fechar a proposta operacional de atualização, backup, retenção, recuperação e monitoramento para o ambiente escolhido e validá-la com a Majucau.
- Verificar os requisitos atuais de suporte do Windows para o ambiente de implantação antes de aprovar um host.

**Progresso registrado:** BK-014 passa de 0% para 10%, apenas pelo inventário parcial do ambiente atual. O critério de aceite não está concluído.
