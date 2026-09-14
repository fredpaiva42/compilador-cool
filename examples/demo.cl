class Animal {
  nome : String;
  patas : Int <- 4;

  init(n : String) : Animal {
    {
      nome <- n;
      self;
    }
  };

  falar() : String { "... " };
  descrever() : String { nome.concat("!") };
};

class Cachorro inherits Animal {
  falar() : String { "Au au!\n" };
};

class Gato inherits Animal {
  falar() : String { "Miau!\n" };
};

class Main inherits IO {
  main() : Object {
    let b : Animal <- new Cachorro,
        g : Gato <- new Gato,
        d : Int <- 8 / 2 * 3,
        e : Int <- 3 - ~2 in
      {
        out_string(b.falar());
        out_string(g@Animal.descrever());
        if isvoid b then abort() else 0 fi;
        if not (1 <= 2) then out_string("zero\n") else out_string("ok\n") fi;
        while false loop out_string("nunca\n") pool;
        case b of
          c : Cachorro => out_string("late\n");
          m : Gato     => out_string("mia\n");
          o : Object   => out_string("outro\n");
        esac;
      }
  };
};
