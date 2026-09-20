"""Exporta os artefatos do planejamento com inventário e hashes; não publica no GitHub."""
import argparse
import hashlib
import json
import zipfile
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', required=True, type=Path, help='Caminho do ZIP a criar, fora da pasta do repositório.')
args = parser.parse_args()
root = Path(__file__).resolve().parents[1]
plan = root / 'docs/planejamento'
dest = args.output.resolve()
if dest.is_relative_to(root):
    parser.error('O ZIP deve ficar fora do repositório para não incluir a si próprio.')
manifest = plan / 'MANIFESTO.json'
paths = [root / name for name in ['README.md', '.gitignore', '.gitattributes']]
paths += [Path(__file__).resolve()]
paths += [p for p in plan.rglob('*') if p.is_file() and not p.is_symlink()
    and not set(p.parts) & {'private','node_modules','__pycache__','.git'}
    and p.suffix in {'.md','.py','.mjs','.json','.txt','.png','.xlsx'}
    and '.inspect.' not in p.name and p != manifest]
paths = sorted(set(paths))
entries = [{'path':p.relative_to(root).as_posix(), 'bytes':p.stat().st_size,
    'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in paths]
manifest.write_text(json.dumps({'description':'Artefatos preparados para guilhermejacyzin/majucau; não comprova envio remoto.',
    'manifest_self_excluded':True, 'files':entries},ensure_ascii=False,indent=2)+'\n',encoding='utf-8',newline='\n')
paths.append(manifest)
dest.parent.mkdir(parents=True,exist_ok=True)
with zipfile.ZipFile(dest,'w',compression=zipfile.ZIP_DEFLATED) as archive:
    for p in paths:
        archive.write(p, 'majucau/'+p.relative_to(root).as_posix())
with zipfile.ZipFile(dest) as archive:
    assert archive.testzip() is None
print(json.dumps({'files':len(paths),'zip':str(dest),'bytes':dest.stat().st_size,
    'sha256':hashlib.sha256(dest.read_bytes()).hexdigest()},ensure_ascii=False))
