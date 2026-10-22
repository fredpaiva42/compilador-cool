(* SELF_TYPE: retorno de metodo, herdado que devolve a classe concreta,
   copy() de Object e receptor self. *)

class Contador {
  valor : Int <- 0;

  incr() : SELF_TYPE {
    {
      valor <- valor + 1;
      self;
    }
  };

  copiar() : SELF_TYPE { self.copy() };

  valorAtual() : Int { valor };
};

class ContadorDuplo inherits Contador {
  incr() : SELF_TYPE {
    {
      valor <- valor + 2;
      self;
    }
  };
};

class Main inherits IO {
  main() : Object {
    let c : ContadorDuplo <- new ContadorDuplo,
        d : Contador <- c.incr(),
        e : ContadorDuplo <- c.copiar()
    in {
      out_int(c.valorAtual());
      out_int(d.valorAtual());
      out_int(e.valorAtual());
      if isvoid e then out_string("vazio\n") else self.copy().out_string("ok\n") fi;
      out_int(c@Contador.valorAtual());
    }
  };
};
