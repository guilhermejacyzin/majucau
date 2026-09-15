import hashlib,json,re
from pathlib import Path
import openpyxl

work=Path('work')
tasks=json.loads((work/'backlog_v1.json').read_text(encoding='utf-8'))
old=json.loads((work/'backlog.json').read_text(encoding='utf-8'))
decisions=json.loads((work/'decisions_v1.json').read_text(encoding='utf-8'))
revision=json.loads((work/'revision_v1.json').read_text(encoding='utf-8'))
by={t['key']:t for t in tasks}
ids=[t['id'] for t in tasks]
assert ids==[t['id'] for t in old]==[f'BK-{i:03}' for i in range(127)]
assert len({d['id'] for d in decisions})==88
assert all(t['status']==o['status'] and t['progress']==o['progress'] and t['weight']==o['weight'] for t,o in zip(tasks,old))
assert all(t['goal'] and t['action'] and t['accept'] and t['impact'] and len(t['steps'])>=3 for t in tasks)

def ancestors(key,stack=()):
    assert key not in stack,('cycle',stack,key)
    ans=set()
    for dep in by[key]['deps']:
        assert dep in by,dep
        ans.add(dep)
        ans.update(ancestors(dep,stack+(key,)))
    return ans
for key in by: ancestors(key)
required=['usersscreen','parametersscreen','integrationsscreen','reconciliationscreen','receivablesscreen','payablesscreen','cashscreen','drescreen','trialbalancescreen','calculatorscreen','applicationsscreen','pricing','logisticsanalysis','inventoryscreen','productionscreen','purchasesscreen','projectionsscreen','projectionexport','pendingscreen','alertsscreen','executivescreen','extrascreens','approvedexports','cyclemetrics']
assert set(required)<=ancestors('readiness')
assert set(['restoreexercise','backupdesign','staginglocal'])<=ancestors('golive')
assert 'supplementaryingest' not in by['investmentledger']['deps']
assert 'supplementaryingest' not in by['cashengine']['deps']
assert 'payableingest' in by['investmentledger']['deps']
assert 'opsingest' in ancestors('inputprojection')
assert all(t['dependencies']==[by[k]['id'] for k in t['deps']] for t in tasks)
for d in decisions:
    if d['id'].startswith(('R-','HT-')): assert d['tasks']
for t in tasks:
    for ref in re.findall(r'\b(?:HT|R|Q|L|C)-\d+\b',t['questions']):
        assert any(d['id']==ref for d in decisions),ref
md=Path('outputs/MAJUCAU_BACKLOG_DETALHADO_v1.md').read_text(encoding='utf-8')
assert re.findall(r'^### (BK-\d+) —',md,flags=re.M)==ids
assert 'Ingerir perdas, ajustes e extratos de investimentos' not in md
assert 'dados confiáveis' in md or 'histórico confiável' in md

path=Path('outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v1.xlsx')
formula=openpyxl.load_workbook(path,data_only=False)
cache=openpyxl.load_workbook(path,data_only=True)
assert formula.sheetnames==['Evolucao','Tarefas','Decisoes','Cobertura']
assert [formula['Tarefas'].cell(i+6,1).value for i in range(127)]==ids
assert cache['Evolucao']['B6'].value==127
assert cache['Evolucao']['B7'].value==3
assert abs(cache['Evolucao']['B8'].value-3/127)<1e-10
assert cache['Evolucao']['B9'].value==0
pending=sum(d['status'] not in ['Respondida','Planejamento autorizado'] for d in decisions)
assert cache['Evolucao']['B10'].value==pending
assert formula['Decisoes']['I5'].value=='Natureza'
assert len(formula['Tarefas'].data_validations.dataValidation)>=2
assert all(formula['Tarefas'].cell(i+6,7).data_type=='f' for i in range(127))
for i,t in enumerate(tasks,6):
    assert formula['Tarefas'].cell(i,3).value==t['title']
    assert formula['Tarefas'].cell(i,10).value==(', '.join(t['dependencies']) or None)
for i,d in enumerate(decisions,6):
    assert formula['Decisoes'].cell(i,1).value==d['id']
    assert formula['Decisoes'].cell(i,3).value==d['status']
    assert formula['Decisoes'].cell(i,9).value==d['nature']
errors=[(s.title,c.coordinate,c.value) for s in cache for row in s for c in row if c.data_type=='e']
assert not errors,errors

manifest=json.loads((work/'source_extract/manifest.json').read_text(encoding='utf-8'))
source_root=Path(r'sources/private/material-original')
assert all(hashlib.sha256((source_root/m['name']).read_bytes()).hexdigest()==m['sha256'] for m in manifest)
assert hashlib.sha256(Path(revision['source']).read_bytes()).hexdigest()==revision['sha256']
assert all(Path('outputs/'+f'MAJUCAU_{stem}_v0.{suffix}').exists() for stem,suffix in [('BACKLOG_DETALHADO','md'),('ANALISE_E_DECISOES','md'),('BACKLOG_E_EVOLUCAO','xlsx')])
result=dict(tasks=127,changed_tasks=len(revision['changes']),stable_ids=True,acyclic_dependencies=True,decisions=len(decisions),pending_records=pending,completed_planning=3,development_progress=0,source_files_unchanged=len(manifest)+1,formula_errors=errors,validation='Document structure, task/decision relationships, dependencies, hashes, XLSX formula caches and visual previews. No Majucau implementation or real integration tests executed.')
(work/'verification_v1.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(result,ensure_ascii=False))
