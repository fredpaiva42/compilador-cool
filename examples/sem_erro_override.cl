(* Erro esperado: override muda a assinatura (Animal.falar nao tem formais).
   Saida: erro semantico: método falar de Cachorro redefine Animal.falar
   com assinatura diferente (linha 10) *)

class Animal {
  falar() : String { "..." };
};

class Cachorro inherits Animal {
  falar(x : Int) : String { "Au au" };
};

class Main inherits IO {
  main() : Object { (new Cachorro).falar() };
};
