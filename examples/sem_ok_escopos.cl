(* Escopos: let encadeado, sombra de atributo, case com join dos ramos,
   while e bloco. *)

class Animal {
  nome : String;
  falar() : String { "..." };
};

class Cachorro inherits Animal {
  falar() : String { "Au au" };
};

class Gato inherits Animal {
  falar() : String { "Miau" };
};

class Main inherits IO {
  nome : String <- "atributo";

  main() : Object {
    let nome : String <- "local",
        a : Animal <- new Cachorro,
        n : Int <- 0
    in {
      out_string(nome);
      case a of
        c : Cachorro => out_string(c.falar());
        g : Gato     => out_string(g.falar());
        o : Animal   => out_string(o.falar());
      esac;
      while n < 3 loop
        {
          out_int(n);
          n <- n + 1;
        }
      pool;
      let x : Int <- 10, y : Int <- x + 5 in out_int(y);
      let n : String <- "sombra" in out_string(n);
      if nome = "local" then out_string("escopo ok\n") else out_string("erro\n") fi;
    }
  };
};
