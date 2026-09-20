from pathlib import Path

script = Path('work/build_tracker.mjs').read_text(encoding='utf-8')
changes = [
    ('work/backlog.json','work/backlog_v1.json'),
    ('work/decisions.json','work/decisions_v1.json'),
    ('_v0.', '_v1.'),
    ('Versão 0 |','Versão 1 |'),
    ('work/previews','work/previews_v1'),
    ("baseStyle(dec,`A1:H${dl}`);widths(dec,{A:75,B:185,C:185,D:560,E:340,F:480,G:275,H:165});", "baseStyle(dec,`A1:I${dl}`);widths(dec,{A:75,B:185,C:185,D:560,E:340,F:480,G:275,H:230,I:255});"),
    ("title(dec,'A2','Majucau — decisões e perguntas');", "title(dec,'A2','Majucau — decisões e verificações');"),
    ("headers(dec,'A5:H5',['ID','Tema','Situação','Pergunta / definição necessária','Impacto','Resposta registrada','Fonte','Tarefas']);", "headers(dec,'A5:I5',['ID','Tema','Situação','Definição / verificação necessária','Impacto','Resposta registrada','Fonte','Tarefas','Natureza']);"),
    ("dec.getRange(`A6:H${dl}`).values=decisions.map(d=>[d.id,d.theme,d.status,d.question,d.impact,d.answer,d.source,d.tasks]);", "dec.getRange(`A6:I${dl}`).values=decisions.map(d=>[d.id,d.theme,d.status,d.question,d.impact,d.answer,d.source,d.tasks,d.nature]);"),
    ("band(dec,6,dl,'H');dec.getRange(`A6:H${dl}`).format.wrapText=true;", "band(dec,6,dl,'I');dec.getRange(`A6:I${dl}`).format.wrapText=true;"),
    ("dec.getRange(`A6:H${dl}`).format.rowHeight=85;", "dec.getRange(`A6:I${dl}`).format.rowHeight=95;"),
    ("['Em aberto','Respondida','Parcialmente esclarecida','Planejamento autorizado']", "['Em aberto','Respondida','Parcialmente esclarecida','Planejamento autorizado','A verificar tecnicamente']"),
    ("dec.tables.add(`A5:H${dl}`,true,'MajucauDecisoes');", "dec.tables.add(`A5:I${dl}`,true,'MajucauDecisoes');"),
    ("['Questões ainda abertas',null,'Sem incluir já respondidas']", "['Registros pendentes',null,'Inclui verificações técnicas']"),
    ('+COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"Parcialmente esclarecida")`', '+COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"Parcialmente esclarecida")+COUNTIFS(\'Decisoes\'!$C$6:$C$${dl},"A verificar tecnicamente")`'),
    ("summary.getRange('A31').values=[['Decisões adicionais: conversa atual. GitHub consultado em modo leitura; repositório vazio.']];", "summary.getRange('A31').values=[['Atualização: consolidação enviada em 13/09/2026. Bling + três estruturas do Drive somente de leitura.']];\nsummary.getRange('A33').values=[['52 tarefas revisadas; mesmos 127 IDs. Integrações e homologações técnicas ainda não executadas.']];\nsummary.getRange('A34').values=[['L/Q e HT podem tratar a mesma frente: registros pendentes não são perguntas distintas.']];\nsummary.getRange('A35').values=[['Resgates e posição dependem de evidência do Bling; rendimento não é aporte de principal.']];"),
    ("['Evolucao','A1:G31']", "['Evolucao','A1:G35']"),
    ("['Decisoes','A1:F9']", "['Decisoes','A69:F74'],['Verificacoes','A84:F88']"),
    ("const png=await wb.render({sheetName:name,range,scale:1,format:'png'});", "const png=await wb.render({sheetName:name==='Verificacoes'?'Decisoes':name,range,scale:1,format:'png'});"),
]
for old,new in changes:
    assert old in script,old
    script=script.replace(old,new)
Path('work/build_tracker_v1.mjs').write_text(script,encoding='utf-8')
print('Gerador v1 preparado; gerador e entregas v0 preservados.')
