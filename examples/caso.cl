class A {
};
class B inherits A {
};
class Main {
  main() : Int {
    { case new B of x : A => 1; y : B => 2; esac
      + case new B of y : B => 20; x : A => 10; esac; }
  };
};
