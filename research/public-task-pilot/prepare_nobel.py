"""New public award facts; corpus/oracle separation, no model access."""
import json
from pathlib import Path

root=Path(__file__).parent
items=[
('physics',2023,'methods producing attosecond light pulses to study electron dynamics',
 'Which award recognized methods producing attosecond light pulses to study electron dynamics?',
 'Which award recognized ways to watch electrons move using extremely brief flashes?'),
('physics',2022,'experiments with entangled photons that violated Bell inequalities and advanced quantum information science',
 'Which award recognized experiments with entangled photons and violations of Bell inequalities?',
 'Which prize honored experiments showing that linked light particles violate the limits imposed by local hidden variables?'),
('chemistry',2023,'the discovery and synthesis of quantum dots',
 'Which award recognized the discovery and synthesis of quantum dots?',
 'Which prize honored making tiny semiconductor particles whose properties depend on their size?'),
('chemistry',2022,'the development of click chemistry and bioorthogonal chemistry',
 'Which award recognized click chemistry and bioorthogonal chemistry?',
 'Which prize honored methods for joining molecular building blocks and reactions compatible with living systems?'),
('medicine',2023,'nucleoside base modifications enabling effective mRNA vaccines against COVID-19',
 'Which award recognized nucleoside base modifications enabling mRNA vaccines against COVID-19?',
 'Which prize honored changes to RNA building blocks that made pandemic vaccination possible?'),
('medicine',2022,'discoveries about genomes of extinct hominins and human evolution',
 'Which award recognized discoveries about genomes of extinct hominins and human evolution?',
 'Which prize honored reading DNA from vanished human relatives to understand our ancestry?')]
corpus=json.loads((root/'prepared-v2/corpus.json').read_text());queries=[];oracle={}
for field,year,topic,literal,paraphrase in items:
    key=f'nobel-{field}-{year}'
    name='Physiology or Medicine' if field=='medicine' else field.title()
    corpus.append(dict(fixture_id=key,text=f'The {year} Nobel Prize in {name} recognized {topic}.',
                       source=f'https://www.nobelprize.org/prizes/{field}/{year}/press-release/'))
    for kind,question in [('literal',literal),('paraphrase',paraphrase)]:
        case=f'{key}-{kind}'
        queries.append(dict(case_id=case,split='confirmation',question=question))
        oracle[case]=dict(answer=f'{name} {year}',support=[key],cluster=key,wording=kind)
for i,question in enumerate(['Which retained award record honors gravitational wave detection?',
                            'Which retained award record recognizes the Higgs boson discovery?']):
    key=f'nobel-absent-{i}'
    queries.append(dict(case_id=key,split='confirmation',question=question))
    oracle[key]=dict(answer='UNKNOWN',support=[],cluster=key,wording='absent')
out=root/'nobel-v1';out.mkdir(exist_ok=True)
for name,value in [('corpus',corpus),('queries',queries),('oracle',oracle)]:
    with (out/f'{name}.json').open('x') as f:json.dump(value,f,indent=2);f.write('\n')
print(len(corpus),'facts;',len(queries),'queries')
