class A {
  f() : Int { 7 };
};
class B inherits A {
};
class Main {
  main() : Int { (new B)@A.f() };
};
