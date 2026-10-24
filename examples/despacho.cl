class Animal {
  valor() : Int { 1 };
};
class Cachorro inherits Animal {
  valor() : Int { 2 };
};
class Gato inherits Animal {
};
class Main {
  main() : Int {
    let a : Animal <- new Cachorro,
        b : Animal <- new Gato,
        c : Int <- 5 in {
      a.valor() + b.valor() + c.copy();
    }
  };
};
