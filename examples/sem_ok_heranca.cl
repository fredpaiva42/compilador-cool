(* Heranca em tres niveis: atributo herdado, override e dispatch estatico. *)

class Forma {
  nome : String;

  init(n : String) : Forma {
    {
      nome <- n;
      self;
    }
  };

  descrever() : String { nome };
};

class Poligono inherits Forma {
  lados : Int <- 0;

  comLados(l : Int) : Poligono {
    {
      lados <- l;
      self;
    }
  };

  descrever() : String { nome.concat(" (poligono)") };

  quantosLados() : Int { lados };
};

class Triangulo inherits Poligono {
  descrever() : String { nome.concat(" (triangulo)") };
};

class Main inherits IO {
  main() : Object {
    let f : Forma <- (new Triangulo).init("T"),
        p : Poligono <- new Poligono
    in {
      out_string(f.descrever());
      out_string(f@Forma.descrever());
      out_string(p.comLados(4).descrever());
      out_int(p.quantosLados());
      out_int(f.descrever().length());
    }
  };
};
