import fs from 'node:fs/promises';
import path from 'node:path';
import {Workbook, SpreadsheetFile} from '@oai/artifact-tool';

const base=path.resolve('.');
const tasks=JSON.parse(await fs.readFile(path.join(base,'work/backlog_v2.json'),'utf8'));
const decisions=JSON.parse(await fs.readFile(path.join(base,'work/decisions_v2.json'),'utf8'));
const scope=JSON.parse(await fs.readFile(path.join(base,'work/scope_v2.json'),'utf8'));
const statusDate=new Intl.DateTimeFormat('pt-BR',{timeZone:'America/Sao_Paulo'}).format(new Date());
const statusCounts={
 completed:tasks.filter(t=>t.status==='Concluída').length,
 inProgress:tasks.filter(t=>t.status==='Em andamento').length,
 notStarted:tasks.filter(t=>t.status==='Não iniciada').length,
};
const totalWeight=tasks.reduce((total,t)=>total+t.weight,0);
const progressPercent=tasks.reduce((total,t)=>total+(t.progress??0)*t.weight,0)/totalWeight*100;
const remainingPercent=100-progressPercent;
const percentText=value=>new Intl.NumberFormat('pt-BR',{minimumFractionDigits:1,maximumFractionDigits:1}).format(value);
const wb=Workbook.create();
const summary=wb.worksheets.add('Evolucao');
const backlog=wb.worksheets.add('Tarefas');
const dec=wb.worksheets.add('Decisoes');
const sc=wb.worksheets.add('Cobertura');
const dark='#0B3346',teal='#087E87',ink='#173445',muted='#566C7A',pale='#F0F5F7',amber='#FFF2CC';
function baseStyle(sh,range){
 sh.showGridLines=false;
 sh.getRange(range).format.font={name:'Arial',size:10,color:ink};
 sh.getRange(range).format.verticalAlignment='center';
 sh.getRange(range).format.rowHeight=28;
}
function title(sh,cell,text){sh.getRange(cell).values=[[text]];sh.getRange(cell).format.font={name:'Arial',size:16,bold:true,color:dark};}
function headers(sh,range,vals){sh.getRange(range).values=[vals];sh.getRange(range).format={fill:dark,font:{name:'Arial',size:10,bold:true,color:'#FFFFFF'},horizontalAlignment:'center',verticalAlignment:'center',wrapText:true};sh.getRange(range).format.rowHeight=34;}
function widths(sh,spec){for(const [col,w] of Object.entries(spec))sh.getRange(`${col}:${col}`).format.columnWidthPx=w;}
function band(sh,start,end,last){for(let r=start;r<=end;r++)if((r-start)%2===0)sh.getRange(`A${r}:${last}${r}`).format.fill=pale;}

const n=tasks.length,last=n+5;
baseStyle(backlog,`A1:N${last}`);backlog.tabColor=teal;
widths(backlog,{A:78,B:215,C:385,D:120,E:142,F:60,G:85,H:260,I:230,J:230,K:230,L:420,M:130,N:130});
title(backlog,'A2','Majucau — tarefas do backlog');
backlog.getRange('A3').values=[['Detalhes de objetivo, execução, aceite e impacto: documento MAJUCAU_BACKLOG_DETALHADO_v2.md, pelo ID.']];
backlog.getRange('A3').format.font={name:'Arial',size:10,italic:true,color:muted};
headers(backlog,'A5:N5',['ID','Fase','Tarefa','Natureza','Situação','Peso','Evolução estimada','Evidência de aceite','Responsável sugerido','Dependências','Decisões a acompanhar','Referências','Peso com aceite','Peso da evolução']);
const rows=tasks.map(t=>[t.id,t.phase,t.title,t.kind,t.status,t.weight,t.progress??0,t.evidence,t.owner,t.dependencies.join(', '),t.questions,t.refs||'Necessidade técnica de entrega e operação.',null,null]);
backlog.getRange(`A6:N${last}`).values=rows;
backlog.getRange(`M6:M${last}`).formulas=rows.map((_,i)=>[`=IF(AND(E${i+6}="Concluída",H${i+6}<>""),F${i+6},0)`]);
backlog.getRange(`N6:N${last}`).formulas=rows.map((_,i)=>[`=F${i+6}*G${i+6}`]);
band(backlog,6,last,'N');backlog.getRange(`A6:N${last}`).format.wrapText=true;
backlog.getRange(`A6:N${last}`).format.rowHeight=72;
backlog.getRange(`E6:F${last}`).format.fill=amber;
backlog.getRange(`H6:H${last}`).format.fill=amber;
backlog.getRange(`F6:G${last}`).format.horizontalAlignment='right';
backlog.getRange(`G6:G${last}`).setNumberFormat('0%');
backlog.getRange(`F6:F${last}`).setNumberFormat('0');
backlog.getRange(`E6:E${last}`).dataValidation={rule:{type:'list',values:['Não iniciada','Aguarda decisão','Em andamento','Em validação','Bloqueada','Concluída']}};
backlog.dataValidations.add({range:`F6:F${last}`,rule:{type:'whole',operator:'between',formula1:1,formula2:100}});
backlog.getRange(`A6:N${last}`).conditionalFormats.addCustom('AND($E6="Concluída",$H6="")',{fill:'#FDE4E0',font:{color:'#A12622'}});
backlog.freezePanes.freezeRows(5);backlog.freezePanes.freezeColumns(3);
backlog.tables.add(`A5:N${last}`,true,'MajucauTarefas');

const dl=decisions.length+5;
baseStyle(dec,`A1:I${dl}`);widths(dec,{A:75,B:185,C:185,D:560,E:340,F:480,G:275,H:230,I:255});
title(dec,'A2','Majucau — regras e verificações');
dec.getRange('A3').values=[['Regras funcionais fechadas. Execução técnica e localização de referências são acompanhadas sem reabrir o negócio.']];
headers(dec,'A5:I5',['ID','Tema','Situação','Definição / verificação necessária','Impacto','Resposta registrada','Fonte','Tarefas','Natureza']);
dec.getRange(`A6:I${dl}`).values=decisions.map(d=>[d.id,d.theme,d.status,d.question,d.impact,d.answer,d.source,d.tasks,d.nature]);
band(dec,6,dl,'I');dec.getRange(`A6:I${dl}`).format.wrapText=true;
dec.getRange(`A6:I${dl}`).format.rowHeight=95;
dec.getRange(`C6:C${dl}`).format.fill=amber;dec.getRange(`F6:F${dl}`).format.fill=amber;
dec.getRange(`C6:C${dl}`).dataValidation={rule:{type:'list',values:['Definida','Execução técnica pendente','Referência a localizar']}};
dec.freezePanes.freezeRows(5);dec.freezePanes.freezeColumns(2);dec.tables.add(`A5:I${dl}`,true,'MajucauDecisoes');

const sl=scope.length+5;
baseStyle(sc,`A1:E${sl}`);widths(sc,{A:180,B:295,C:305,D:560,E:125});
title(sc,'A2','Majucau — cobertura dos módulos');
sc.getRange('A3').values=[['Balanço Patrimonial foi excluído por decisão atual. Itens incompletos permanecem para refinamento.']];
headers(sc,'A5:E5',['Identificação','Módulo','Material disponível','Referência','Tarefa de interface']);
sc.getRange(`A6:E${sl}`).values=scope.map(s=>[s.code,s.module,s.coverage,s.source,s.task]);
band(sc,6,sl,'E');sc.getRange(`A6:E${sl}`).format.wrapText=true;sc.getRange(`A6:E${sl}`).format.rowHeight=54;
sc.freezePanes.freezeRows(5);sc.tables.add(`A5:E${sl}`,true,'MajucauCobertura');

baseStyle(summary,'A1:G43');widths(summary,{A:295,B:125,C:170,D:28,E:340,F:150,G:150});summary.tabColor=dark;
title(summary,'A2','Majucau — evolução do backlog');
summary.getRange('A3').values=[[`Status atualizado em ${statusDate} | fechamento funcional de 13/09/2026`]];
summary.getRange('A3').format.font={name:'Arial',size:10,italic:true,color:muted};
headers(summary,'A5:C5',['Medida','Valor','Interpretação']);
const metrics=[['Tarefas no catálogo',n,'Escopo provisório'],['Levantamentos concluídos',null,'Com evidência de aceite'],['Evolução estimada por peso',null,'Estimativa; não é prontidão'],['Desenvolvimento aceito',null,'Somente concluído com evidência'],['Frentes técnicas',null,'HT: execução futura']];
summary.getRange('A6:C10').values=metrics;
summary.getRange('B7').formulas=[[`=SUMIFS('Tarefas'!$M$6:$M$${last},'Tarefas'!$D$6:$D$${last},"Planejamento")`]];
summary.getRange('B8').formulas=[[`=SUM('Tarefas'!$N$6:$N$${last})/SUM('Tarefas'!$F$6:$F$${last})`]];
summary.getRange('B9').formulas=[[`=SUMIFS('Tarefas'!$M$6:$M$${last},'Tarefas'!$D$6:$D$${last},"Desenvolvimento")/SUMIFS('Tarefas'!$F$6:$F$${last},'Tarefas'!$D$6:$D$${last},"Desenvolvimento")`]];
summary.getRange('B10').formulas=[[`=COUNTIFS('Decisoes'!$I$6:$I$${dl},"Homologação técnica")`]];
summary.getRange('A11:C11').values=[['Tarefas aceitas com evidência',null,'Contagem de entregas concluídas']];
summary.getRange('B11').formulas=[[`=COUNTIFS('Tarefas'!$E$6:$E$${last},"Concluída",'Tarefas'!$H$6:$H$${last},"<>")`]];
summary.getRange('B8:B9').setNumberFormat('0.0%');
summary.getRange('A6:C11').format.rowHeight=36;summary.getRange('C6:C11').format.wrapText=true;
summary.getRange('B6:B11').format.horizontalAlignment='right';
summary.getRange('B6:B11').format.font={name:'Arial',size:14,bold:true,color:teal};
summary.getRange('A12').values=[['Avanço por fase']];summary.getRange('A12').format.font={bold:true,color:dark};
headers(summary,'A13:C13',['Fase','Tarefas','Evolução estimada']);
const phases=[...new Set(tasks.map(t=>t.phase))];
summary.getRange(`A14:A${13+phases.length}`).values=phases.map(p=>[p]);
summary.getRange(`B14:B${13+phases.length}`).formulas=phases.map((_,i)=>[`=COUNTIFS('Tarefas'!$B$6:$B$${last},A${i+14})`]);
summary.getRange(`C14:C${13+phases.length}`).formulas=phases.map((_,i)=>[`=SUMIFS('Tarefas'!$N$6:$N$${last},'Tarefas'!$B$6:$B$${last},A${i+14})/SUMIFS('Tarefas'!$F$6:$F$${last},'Tarefas'!$B$6:$B$${last},A${i+14})`]);
summary.getRange(`C14:C${13+phases.length}`).setNumberFormat('0%');band(summary,14,13+phases.length,'C');
summary.getRange(`A14:C${13+phases.length}`).format.rowHeight=30;
headers(summary,'E5:G5',['Decisões já recebidas','','']);
const fixed=[['Empresa','Somente Majucau'],['Usuários','3 diretores'],['Hospedagem','Atual local; destino privado centralizado'],['Primeira produção','Todos os módulos confirmados'],['Pagamento diferente','Divergência, sem baixa'],['Caixa da aplicação','Seguir estrutura enviada; fluxo consolidado D0–D+60'],['DRE','Calculada pelo Majucau; Bling para comparação'],['Cards EBIT/EBITDA','No espaço aprovado no lugar de UNMAPPED'],['Cards CAPEX','Pago no mês e acumulado; fora do lucro'],['Balanço Patrimonial','Fora do escopo'],['Referências','Última versão aprovada']];
summary.getRange('E6:F16').values=fixed;
summary.getRange('F6:F16').format.columnWidthPx=300;
summary.getRange('E6:F16').format.rowHeight=30;
summary.getRange('E17').values=[['Como atualizar']];summary.getRange('E17').format.font={bold:true,color:dark};
const instructions=[
'Na aba Tarefas, atualize Situação, Evolução estimada e Evidência.',
'Concluída com evidência conta como aceite comprovado.',
'A evolução estimada aparece antes do aceite final.',
'A média global é estimativa do backlog; não é prontidão para produção.',
'Peso inicial 1 mede quantidade, sem estimativa de esforço.',
'Após estimar com a equipe, revise os pesos e registre a mudança.',
'As fórmulas consideram todas as linhas atuais, mesmo filtradas.',
'Ao incluir tarefas, ampliar tabela e fórmulas de resumo.',
'Detalhamento passo a passo fica no documento pelo ID BK.',
'Não existe atualização automática de atividade no GitHub.'
];
summary.getRange('E18:E27').values=instructions.map(s=>[s]);
summary.getRange('E18:E27').format.rowHeight=29;
summary.getRange('A28').values=[['Responsáveis são papéis sugeridos; prazo e esforço aguardam a equipe.']];
summary.getRange('A30').values=[['Fonte: imagens e planilha MAJUCAU_CONSOLIDACAO_DEV_GATE_RC3_CEI_CONFIGURACOES_T16_T17.xlsx fornecidas pelo usuário.']];
summary.getRange('A31').values=[['Fechamento de 13/09/2026; decisões posteriores registradas no G0-DECISION-REGISTER.md até 11/10/2026.']];
summary.getRange('A33').values=[[`Backlog em ${statusDate}: ${statusCounts.completed} concluídas, ${statusCounts.inProgress} em andamento e ${statusCounts.notStarted} não iniciadas. Estimativa: ${percentText(progressPercent)}%; restante: ${percentText(remainingPercent)}%. É uma estimativa, não prova de prontidão.`]];
summary.getRange('A34').values=[['D-1: análise/conciliação normal até ontem; agenda futura continua na projeção.']];
summary.getRange('A35').values=[['Nenhuma regra foi reaberta nesta auditoria. Paginação do Bling e outras pendências técnicas continuam em verificação.']];
summary.getRange('A30:A31').format.font={name:'Arial',size:10,italic:true,color:muted};

// Verifica transição de andamento/conclusão sem confundir falta de evidência.
const testRow=10;
const savedStatus=backlog.getRange(`E${testRow}`).values;
const savedEvidence=backlog.getRange(`H${testRow}`).values;
backlog.getRange(`E${testRow}`).values=[['Concluída']];
backlog.getRange(`H${testRow}`).values=[['']];
if(backlog.getRange(`M${testRow}`).values[0][0]!==0)throw new Error('Conclusão sem evidência não pode contar como aceite');
backlog.getRange(`H${testRow}`).values=[['Evidência temporária para teste do rastreador']];
if(backlog.getRange(`M${testRow}`).values[0][0]!==1)throw new Error('Conclusão com evidência deve contar como aceite');
backlog.getRange(`E${testRow}`).values=savedStatus;backlog.getRange(`H${testRow}`).values=savedEvidence;
wb.recalculate();
const view=await wb.inspect({kind:'table',range:'Evolucao!A5:C11',include:'values,formulas',tableMaxRows:7,tableMaxCols:3,maxChars:2500});
console.log(view.ndjson);
const errors=await wb.inspect({kind:'match',searchTerm:'#REF!|#DIV/0!|#VALUE!|#NAME\\?|#N/A|#NUM!|#NULL!',options:{useRegex:true,maxResults:20},maxChars:1000});console.log(errors.ndjson);
const previewDir=path.resolve(base,'..','..','artifacts','verification','backlog-v2');
await fs.mkdir(previewDir,{recursive:true});
for(const [name,range] of [['Evolucao','A1:G35'],['Tarefas','A1:N12'],['Decisoes','A94:F99'],['Verificacoes','A84:F88'],['Cobertura','A1:E11']]){
 const png=await wb.render({sheetName:name==='Verificacoes'?'Decisoes':name,range,scale:1,format:'png'});
 await fs.writeFile(path.join(previewDir,name+'.png'),new Uint8Array(await png.arrayBuffer()));
}
const file=await SpreadsheetFile.exportXlsx(wb);
await file.save(path.join(base,'outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx'));
// O artifact-tool grava a última inspeção ao lado da planilha. Mova-a para
// artifacts/ para manter as saídas temporárias centralizadas fora de outputs/.
const inspectOutput=path.join(base,'outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx.inspect.ndjson');
const inspectArchive=path.join(previewDir,'MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx.inspect.ndjson');
if(await fs.access(inspectOutput).then(()=>true).catch(()=>false)){
 await fs.rm(inspectArchive,{force:true});
 await fs.rename(inspectOutput,inspectArchive);
}
console.log(JSON.stringify({tasks:n,decisions:decisions.length,summary:summary.getRange('B6:B10').values,output:'MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx'}));
