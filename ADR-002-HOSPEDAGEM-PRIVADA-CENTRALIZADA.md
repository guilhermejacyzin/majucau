# ADR-002 — Hospedagem privada centralizada

- **Status:** aprovada como direção-alvo pela responsável pelo produto em 2026-10-03; implementação pendente
- **Substitui:** ADR-001 somente quanto ao destino local single-user do PostgreSQL e à operação multi-site; não altera telas nem regras de negócio aprovadas
- **Contexto:** ingestão diária de dados, operação em mais de um local e consulta por duas pessoas

## Decisão

O destino do produto será uma implantação hospedada e centralizada, com acesso autenticado pela camada de aplicação. PostgreSQL permanecerá em uma rede privada, acessível somente pelos serviços confiáveis do Majucau. O aplicativo cliente não receberá credenciais de banco nem fará conexões diretas ao PostgreSQL.

As pessoas terão identidades individuais. O perfil de consulta começa com privilégio de leitura; permissões de gravação ou administração exigem decisões específicas. O serviço central será responsável por sincronização diária, idempotência, auditoria e persistência. A seleção do provedor, orçamento, região dos dados e produto de identidade ainda não foi aprovada.

A prévia de CSV permanece local e sem dependência do banco. Antes de gravar um arquivo no repositório central, o serviço deverá validar novamente o conteúdo e aplicar as regras já aprovadas. O canal de envio, retenção do arquivo e experiência de importação ainda precisam de desenho e aceite.

## Segurança e operação obrigatórias

- somente a API/serviço autenticado aceita tráfego de clientes; TLS protege as conexões;
- PostgreSQL não possui endpoint público e aceita conexão somente do serviço autorizado;
- segredos de integração ficam em armazenamento de segredos apropriado ao host, sem migração automática de tokens DPAPI;
- usuários individuais, privilégio mínimo, trilha de auditoria e separação dos ambientes de desenvolvimento e produção;
- backups cifrados, retenção definida, monitoramento e ensaio de restauração antes do gate de produção;
- retry, atomicidade, idempotência, lineage e critérios financeiros aprovados permanecem inalterados.

## Transição

O código atual continua sendo uma aplicação Wails/worker com PostgreSQL local. Essa configuração é a base de desenvolvimento e não deve ser anunciada como implantação multi-site. A migração será planejada como G8 e não poderá trocar o mecanismo de armazenamento de credenciais ou expor dados até que o modelo de ameaças, identidade, acesso, backup/restore e estratégia de ingestão estejam aprovados.

As telas aprovadas permanecem como estão. Uma mudança de apresentação ou regra de negócio exige aprovação separada.

## Decisões ainda necessárias

1. provedor, orçamento, região e requisitos de disponibilidade;
2. acesso à aplicação pela internet com autenticação forte ou rede privada/VPN;
3. forma final do cliente: Wails conectado ao serviço ou aplicação em navegador;
4. onde e como a sincronização diária roda, com alertas e recuperação de atraso;
5. canal seguro para os CSVs do local externo, retenção e eliminação dos temporários;
6. identidade, MFA, perfis, auditoria e prazo de retenção dos dados.

## Referências de segurança

- [PostgreSQL — conexões TCP/IP com TLS](https://www.postgresql.org/docs/current/ssl-tcp.html)
- [PostgreSQL — interfaces de escuta e autenticação](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [AWS RDS — criptografia de dados e backups como exemplo de controle de serviço gerenciado](https://docs.aws.amazon.com/AmazonRDS/latest/gettingstartedguide/advanced-security.html)
