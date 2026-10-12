#!/usr/bin/env python3
"""Confere a estrutura de parenteses de um arquivo .wat.

Ignora strings e comentarios de linha (;;).
Uso: python3 check.py arquivo.wat
"""
import re
import sys

linhas = open(sys.argv[1], encoding="utf-8").read().split("\n")
depth = 0
fechou_em = None
for i, linha in enumerate(linhas, 1):
    l = re.sub(r";;[^\n]*", "", linha)
    l = re.sub(r'"[^"\n]*"', '""', l)
    for c in l:
        if c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
            if depth < 0:
                print(f"ERRO: parentese de fechamento a mais na linha {i}")
                sys.exit(1)
    if depth == 0 and fechou_em is None and l.strip():
        fechou_em = i

ultima = max(i for i, linha in enumerate(linhas, 1) if linha.strip())
if depth != 0:
    print(f"ERRO: faltam {depth} parentese(s) de fechamento")
    sys.exit(1)
if fechou_em is not None and fechou_em < ultima:
    print(f"ALERTA: a estrutura fecha na linha {fechou_em} e ainda ha texto ate a linha {ultima}")
    print("        parentese a mais no meio: o modulo fecha cedo e o que vem depois fica fora")
    sys.exit(1)
print("ok: estrutura consistente")
