class Main {
  main() : Int {
    let n : Int <- 5, acc : Int <- 1 in {
      while 1 < n loop {
        acc <- acc * n;
        n <- n - 1;
      } pool;
      acc;
    }
  };
};
