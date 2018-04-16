DROP TABLE IF EXISTS curso;
CREATE TABLE curso (
    curso_id integer PRIMARY KEY AUTOINCREMENT,
    nome text
);

DROP TABLE IF EXISTS participante;
CREATE TABLE participante (
    participante_id integer PRIMARY KEY AUTOINCREMENT,
    curso_id integer,
    nome text,
    instituicao text,
    FOREIGN KEY (curso_id) REFERENCES curso(curso_id)
    ON DELETE CASCADE ON UPDATE NO ACTION
);

DROP TABLE IF EXISTS minicurso;
CREATE TABLE minicurso (
    minicurso_id integer PRIMARY KEY AUTOINCREMENT,
    nome text,
    palestrante text,
    horario_comeco datetime,
    horario_fim datetime,
    vagas integer,
    vagas_restantes integer
);

DROP TABLE IF EXISTS participante_minicurso;
CREATE TABLE participante_minicurso (
    participante_id integer,
    minicurso_id integer,
    PRIMARY KEY (participante_id, minicurso_id),
    FOREIGN KEY (participante_id) REFERENCES participante (participante_id)
    ON DELETE CASCADE ON UPDATE NO ACTION,
    FOREIGN KEY (minicurso_id) REFERENCES minicurso (minicurso_id)
    ON DELETE CASCADE ON UPDATE NO ACTION
);

INSERT INTO curso(nome) values  ('Zootecnia'),
                                ('Eng. Agronômica'),
                                ('Tecnologia em Alimentos'),
                                ('Lic. Química'),
                                ('Lic. Ciências Biológicas');

INSERT INTO minicurso(nome, palestrante, horario_comeco, horario_fim, vagas, vagas_restantes) values
    ('Prática comportamental em cães: avaliando situações de risco e reduzindo o estresse durante as manipulações', 'Ana Paula Ribeiro', '2018-05-16 07:30', '2018-05-16 11:30', 25, 25),
    ('Leite: composição x manejo', 'Bento José Ribeiro', '2018-05-16 07:30', '2018-05-16 11:30', 30, 30),
    ('Ferramenta prática para formulação de ração (Turma 1)', 'Flávio Salvador', '2018-05-16 07:30', '2018-05-16 11:30', 20, 20),
    ('Boas práticas na colheita, extração e beneficiamento do mel', 'José Antônio Bessa', '2018-05-16 07:30', '2018-05-16 11:30', 25, 25),
    ('Prática comportamental em cães: avaliando situações de risco e reduzindo o estresse durante as manipulações', 'Ana Paula Ribeiro', '2018-05-16 12:30', '2018-05-16 16:30', 25, 25),
    ('Manejo reprodutivo de éguas', '', '2018-05-16 12:30', '2018-05-16 16:30', 25, 25),
    ('Ferramenta prática para formulação de ração (Turma 2)', 'Flávio Salvador', '2018-05-16 12:30', '2018-05-16 16:30', 20, 20),
    ('Inseminação artificial em perus', 'Francisco Ailton Batista - BRFoods', '2018-05-16 12:30', '2018-05-16 16:30', 25, 25),
    ('Produção de derivados lácteos', 'Marlene Jerônimo', '2018-05-17 07:30', '2018-05-17 11:30', 20, 20),
    ('Manejo geral na criação de ovinos', 'Sarita Bonagurio Gallo', '2018-05-17 07:30', '2018-05-17 11:30', 25, 25),
    ('Manejo de animais silvestres', 'Lucas Andrade Carneiro- diretor de alimentação e nutrição de animais no zoológico de Brasília', '2018-05-17 07:30', '2018-05-17 11:30', 25, 25),
    ('Sistema de suínos criados ao ar livre – SISCAL', '– professor na Faculdades Associadas de Uberaba (FAZU) e Universidade de Uberaba (UNIUBE)', '2018-05-17 07:30', '2018-05-17 11:30', 25, 25),
    ('Produção de derivados lácteos', 'Marlene Jerônimo', '2018-05-17 12:30', '2018-05-17 16:30', 20, 20),
    ('Manejo geral na criação de ovinos', 'Sarita Bonagurio Gallo', '2018-05-17 12:30', '2018-05-17 16:30', 25, 25),
    ('Avaliação da qualidade de carnes', 'Lucas Arantes Pereira', '2018-05-17 12:30', '2018-05-17 16:30', 25, 25),
    ('Forragicultura', '', '2018-05-17 12:30', '2018-05-17 16:30', 25, 25);