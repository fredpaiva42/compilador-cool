(* Erro esperado: atributo herdado nao pode ser redefinido.
   Saida: erro semantico: atributo nome de Cachorro redefine herdado (linha 9) *)

class Animal {
  nome : String;
};

class Cachorro inherits Animal {
  nome : String <- "Rex";
};

class Main inherits IO {
  main() : Object { new Cachorro };
};
