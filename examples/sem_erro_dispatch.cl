(* Erro esperado: latir() so existe em Cachorro, e o tipo estatico de a
   e Animal; a verificacao e pelo tipo declarado, nao pelo objeto em runtime.
   Saida: erro semantico: método latir inexistente em Animal (linha 15) *)

class Animal {
  comer() : Object { self };
};

class Cachorro inherits Animal {
  latir() : String { "Au au" };
};

class Main inherits IO {
  main() : Object {
    let a : Animal <- new Cachorro in a.latir()
  };
};
