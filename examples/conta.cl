class Main inherits IO {
  main() : Object {
    let s : String <- "ola, " in {
      out_string(s.concat("compilador").substr(5, 10));
    }
  };
};
