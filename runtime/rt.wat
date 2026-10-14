;; Runtime COOL -> WASM. Convenções de memória:
;; 0..7 iovec de escrita {ptr, len}
;; 8..11 nwritten
;; 16..23 iovec de leitura {ptr, len}
;; 24..27 nread
;; 32..63   buffer de conversão (putint escreve de trás para frente)
;; 64..1087 buffer de leitura (1024 bytes)
;; 2048 strings literais (segmentos de dados)
;; 65536.. heap (objetos), cresce para cima
;;
;; Tipos: valores de Int são i64; endereços, tags e booleanos são i32.

(module
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))
  (import "wasi_snapshot_preview1" "fd_read" (func $fd_read (param i32 i32 i32 i32) (result i32)))

  (memory (export "memory") 16)

  ;; próximo endereço livre do heap
  (global $heap (mut i32)
    (i32.const 65536))
  (global $in_pos (mut i32)
    (i32.const 0))
  (global $in_len (mut i32)
    (i32.const 0))

  ;; quebra de linha, para separar as saídas do teste
  (data
    (i32.const 2048) "\n")
  (data (i32.const 2050) "linha: ")
  (data (i32.const 2060) "primeiro: ")
  (data (i32.const 2072) "segundo: ")
  (data (i32.const 2084) "neg: ")

  ;; putstr(ptr, len): escreve len bytes a partir de ptr
  (func $putstr (param $ptr i32) (param $len i32)
    (i32.store
      (i32.const 0)
      (local.get $ptr))
    (i32.store
      (i32.const 4)
      (local.get $len))
    (drop
      (call $fd_write
        (i32.const 1)
        (i32.const 0)
        (i32.const 1)
        (i32.const 8))))

  (func (export "_start")
    (local $n i32)
    (local.set $n (call $read_line (i32.const 1100) (i32.const 200)))
    (call $putstr (i32.const 2050) (i32.const 7))
    (call $putstr (i32.const 1100) (local.get $n))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putstr (i32.const 2060) (i32.const 10))
    (call $putint (call $read_int))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putstr (i32.const 2072) (i32.const 9))
    (call $putint (call $read_int))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putstr (i32.const 2084) (i32.const 5))
    (call $putint (call $read_int))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $proc_exit (i32.const 0)))


  ;; alloc(size): devolve o próximo endereço livre e avança o heap
  ;; O tamanho e arrendondado para múltiplo de 8: (size + 7) & -8
  (func $alloc (param $size i32) (result i32)
    (local $p i32)
    (local.set $p
      (global.get $heap))
    (global.set $heap
      (i32.add
        (local.get $p)
        (i32.and
          (i32.add
            (local.get $size)
            (i32.const 7))
          (i32.const -8))))
    (local.get $p))

  ;; putint(n): imprime n em decimal.
  ;; Escreve os dígitos de tras para frente no buffer 32..63 e imprime
  ;; o pedaço usado de uma vez só.
  (func $putint (param $n i64)
    (local $p i32)
    (local $neg i32)
    (local $d i64)
    (local.set $p
      (i32.const 64))
    (if
      (i64.lt_s
        (local.get $n)
        (i64.const 0))
      (then
        (local.set $neg
          (i32.const 1))
        (local.set $n
          (i64.sub
            (i64.const 0)
            (local.get $n)))))
    (block $fim
      (loop $volta
        (local.set $d
          (i64.rem_u
            (local.get $n)
            (i64.const 10)))
        (local.set $n
          (i64.div_u
            (local.get $n)
            (i64.const 10)))
        (local.set $p
          (i32.sub
            (local.get $p)
            (i32.const 1)))
        (i32.store8
          (local.get $p)
          (i32.add
            (i32.const 48)
            (i32.wrap_i64
              (local.get $d))))
        (br_if $fim
          (i64.eqz
            (local.get $n)))
        (br $volta)))
    (if
      (local.get $neg)
      (then
        (local.set $p
          (i32.sub
            (local.get $p)
            (i32.const 1)))
        (i32.store8
          (local.get $p)
          (i32.const 45))))
    (call $putstr
      (local.get $p)
      (i32.sub
        (i32.const 64)
        (local.get $p))))

  ;; refil: traz bytes da entrada padrão para o buffer 64..1087
  (func $refill
    (i32.store
      (i32.const 16)
      (i32.const 64))
    (i32.store
      (i32.const 20)
      (i32.const 1024))
    (drop
      (call $fd_read
        (i32.const 0)
        (i32.const 16)
        (i32.const 1)
        (i32.const 24)))
    (global.set $in_len
      (i32.load
        (i32.const 24)))
    (global.set $in_pos
      (i32.const 0)))

  ;; peek: olha o próximo byte sem consumir; -1 no fim da entrada
  (func $peek (result i32)
    (if
      (i32.ge_u
        (global.get $in_pos)
        (global.get $in_len))
      (then
        (call $refill)
        (if
          (i32.eqz
            (global.get $in_len))
          (then
            (return
              (i32.const -1))))))
    (i32.load8_u
      (i32.add
        (i32.const 64)
        (global.get $in_pos))))

  ;; advance: consome o byte atual
  (func $advance
    (global.set $in_pos
      (i32.add
        (global.get $in_pos)
        (i32.const 1))))

  ;; is_space: espaço, nova linha, tabulação ou retorno de carro
  (func $is_space (param $c i32) (result i32)
    (i32.or
      (i32.eq
        (local.get $c)
        (i32.const 32))
      (i32.or
        (i32.eq
          (local.get $c)
          (i32.const 10))
        (i32.or
          (i32.eq
            (local.get $c)
            (i32.const 9))
          (i32.eq
            (local.get $c)
            (i32.const 13))))))

  ;; read_line(dst, max): lê uma linha sem o '\n' e devolve o tamanho
  (func $read_line (param $dst i32) (param $max i32) (result i32)
    (local $len i32)
    (local $c i32)
    (block $fim
      (loop $volta
        (local.set $c (call $peek))
        (br_if $fim (i32.eq (local.get $c) (i32.const -1)))
        (call $advance)
        (br_if $fim (i32.eq (local.get $c) (i32.const 10)))
        (if (i32.lt_u (local.get $len) (local.get $max))
          (then
            (i32.store8 (i32.add (local.get $dst) (local.get $len)) (local.get $c))
            (local.set $len (i32.add (local.get $len) (i32.const 1)))
          )
        )
        (br $volta)
      )
    )
    (local.get $len)
  )

  ;; skip_line consome até o fim da linha (inclusive o '\n') ou o fim da entrada
  (func $skip_line
    (local $c i32)
    (block $fim
      (loop $volta
        (local.set $c (call $peek))
        (br_if $fim (i32.eq (local.get $c) (i32.const -1)))
        (call $advance)
        (br_if $fim (i32.eq (local.get $c) (i32.const 10)))
        (br $volta)
      )
    )
  )

  ;; read_int: pula espaços, lê sinal e dígitos, descarta o resto da linha;
  ;; devolve 0 se não houver número
  (func $read_int (result i64)
    (local $c i32)
    (local $neg i32)
    (local $v i64)
    (block $fim_espacos
      (loop $pula
        (local.set $c (call $peek))
        (br_if $fim_espacos (i32.eq (local.get $c) (i32.const -1)))
        (br_if $fim_espacos (i32.eqz (call $is_space (local.get $c))))
        (call $advance)
        (br $pula)
      )
    )
    (local.set $c (call $peek))
    (if (i32.eq (local.get $c) (i32.const 45))
      (then
        (local.set $neg (i32.const 1))
        (call $advance)
      )
    )
    (block $fim_digitos
      (loop $digitos
        (local.set $c (call $peek))
        (br_if $fim_digitos (i32.lt_s (local.get $c) (i32.const 48)))
        (br_if $fim_digitos (i32.gt_s (local.get $c) (i32.const 57)))
        (local.set $v
          (i64.add (i64.mul (local.get $v) (i64.const 10))
            (i64.extend_i32_u (i32.sub (local.get $c) (i32.const 48)))
          )
        )
        (call $advance)
        (br $digitos)
      )
    )
    (call $skip_line)
    (if (local.get $neg)
      (then
        (local.set $v (i64.sub (i64.const 0) (local.get $v)))
      )
    )
    (local.get $v)
  )
)
