from pathlib import Path
import json

# O gerador ativo é atualizado; entregas v0/v1 permanecem intactas.
s=Path('work/build_tracker_v1.mjs').read_text(encoding='utf-8')
s=s.replace('_v1','_v2').replace('Versão 1 |','Versão 2 |').replace('work/scope.json','work/scope_v2.json')
n=len(json.loads(Path('work/revision_v2.json').read_text(encoding='utf-8'))['changes'])
replacements=[
    ('Majucau — decisões e verificações','Majucau — regras e verificações'),
    ('Respostas atuais prevalecem apenas nos pontos respondidos; divergências restantes serão decididas caso a caso.','Regras funcionais fechadas. Execução técnica e localização de referências são acompanhadas sem reabrir o negócio.'),
    ("['Em aberto','Respondida','Parcialmente esclarecida','Planejamento autorizado','A verificar tecnicamente']", "['Definida','Execução técnica pendente','Referência a localizar']"),
    ("['Registros pendentes',null,'Inclui verificações técnicas']", "['Frentes técnicas',null,'HT: execução futura']"),
    ('=COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"Em aberto")+COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"Parcialmente esclarecida")+COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"A verificar tecnicamente")', '=COUNTIFS(\'Decisoes\'!$I$6:$I$${dl},"Homologação técnica")'),
    ("['Lucro elegível','DRE Majucau']", "['DRE de origem','Bling por competência']"),
    ("['Regras × imagens','Decisão caso a caso']", "['Referências','Última versão aprovada']"),
    ('52 tarefas revisadas; mesmos 127 IDs. Integrações e homologações técnicas ainda não executadas.',f'{n} tarefas revisadas; 127 IDs preservados. Regra definida não equivale a implementação concluída.'),
    ('L/Q e HT podem tratar a mesma frente: registros pendentes não são perguntas distintas.','D-1: análise/conciliação normal até ontem; agenda futura continua na projeção.'),
    ('Resgates e posição dependem de evidência do Bling; rendimento não é aporte de principal.','Sem novas perguntas de negócio nesta revisão. Fontes reais e referências serão verificadas tecnicamente.'),
    ('Atualização: consolidação enviada em 13/09/2026. Bling + três estruturas do Drive somente de leitura.','Atualização: fechamento vigente de 13/09/2026. Bling + três estruturas do Drive somente de leitura.'),
    ("['Decisoes','A69:F74'],['Verificacoes','A84:F88']", "['Decisoes','A94:F99'],['Verificacoes','A84:F88']"),
]
for a,b in replacements:
    assert a in s,a
    s=s.replace(a,b)
Path('work/build_tracker.mjs').write_text(s,encoding='utf-8')
print('Gerador ativo configurado para versão 2; arquivos anteriores preservados.')
