(* peculiaridades.cl: exercita os 7 pontos especificos de Cool. *)
class Animal inherits IO {
  nome : String;
  falar() : String { "... " };
  descrever() : String { nome.concat("!") };
};

class Cachorro inherits Animal {
  falar() : String { "Au au!\n" };
};

class Main inherits IO {
  main() : Object {
    let b : Animal <- new Cachorro,
        x : Int <- 0,
        y : Int <- 1 in
      {
        (* 3: atribuicao so em identificador, associativa a direita *)
        x <- y <- 2;
        (* 4: dispatch normal com . e chamada local com self implicito *)
        out_string(b.falar());
        (* 4: dispatch estatico com @ *)
        out_string(b@Animal.descrever());
        (* 2: not com precedencia abaixo da comparacao; 1 comparacao simples, nao em cadeia *)
        if not (1 <= 2) then out_string("zero\n") else out_string("ok\n") fi;
        (* 5: if exige else e fi; while exige loop e pool *)
        while false loop out_string("nunca\n") pool;
        case b of
          c : Cachorro => out_string("late\n");
          o : Object => out_string("outro\n");
        esac;
        (* 6: bloco aninhado dentro de metodo *)
        {
          out_string("aninhado\n");
        };
      }
  };
};
