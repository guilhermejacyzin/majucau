import json, hashlib
from pathlib import Path
import openpyxl

src = Path(r'sources/private/material-original')
out = Path('work/source_extract')
out.mkdir(parents=True, exist_ok=True)
book = openpyxl.load_workbook(src / 'MAJUCAU_CONSOLIDACAO_DEV_GATE_RC3_CEI_CONFIGURACOES_T16_T17.xlsx', read_only=True, data_only=False)
all_data = {}
for sheet in book:
    rows = []
    for row in sheet.iter_rows():
        cells = {c.coordinate: c.value for c in row if c.value is not None}
        if cells:
            rows.append(cells)
    all_data[sheet.title] = rows
    text = '\n'.join(' | '.join(f'{k}={v}' for k,v in row.items()) for row in rows)
    (out / (sheet.title + '.txt')).write_text(text, encoding='utf-8')
    print(sheet.title, 'nonempty rows:',len(rows),'characters:',len(text))
(out/'workbook.json').write_text(json.dumps(all_data,ensure_ascii=False,indent=2,default=str),encoding='utf-8')
manifest=[]
for p in sorted(src.iterdir()):
    if p.is_file():
        manifest.append({'name':p.name,'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
(out/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2),encoding='utf-8')
print('Total nonempty rows:',sum(len(r) for r in all_data.values()))
