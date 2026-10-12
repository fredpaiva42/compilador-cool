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

  (memory (export "memory") 16)

  ;; próximo endereço livre do heap
  (global $heap (mut i32) (i32.const 65536))

  ;; quebra de linha, para separar as saídas do teste
  (data (i32.const 2048) "\n")

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
    (call $putint (i64.const 0))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putint (i64.const 42))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putint (i64.const -7))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putint (i64.const 1234567890))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $putint (i64.const 9223372036854775807))
    (call $putstr (i32.const 2048) (i32.const 1))
    (call $proc_exit (i32.const 0)))

  ;; alloc(size): devolve o próximo endereço livre e avança o heap
  ;; O tamanho e arrendondado para múltiplo de 8: (size + 7) & -8
  (func $alloc (param $size i32) (result i32)
    (local $p i32)
    (local.set $p (global.get $heap))
    (global.set $heap
      (i32.add (local.get $p)
        (i32.and (i32.add (local.get $size) (i32.const 7 )) (i32.const -8))))
    (local.get $p))

;; putint(n): imprime n em decimal.
;; Escreve os dígitos de tras para frente no buffer 32..63 e imprime
;; o pedaço usado de uma vez só.
  (func $putint (param $n i64)
    (local $p i32)
    (local $neg i32)
    (local $d i64)
    (local.set $p (i32.const 64))
    (if (i64.lt_s (local.get $n) (i64.const 0))
      (then
        (local.set $neg (i32.const 1))
        (local.set $n (i64.sub (i64.const 0) (local.get $n)))))
    (block $fim
      (loop $volta
        (local.set $d (i64.rem_u (local.get $n) (i64.const 10)))
        (local.set $n (i64.div_u (local.get $n) (i64.const 10)))
        (local.set $p (i32.sub (local.get $p) (i32.const 1)))
        (i32.store8 (local.get $p)
          (i32.add (i32.const 48) (i32.wrap_i64 (local.get $d))))
        (br_if $fim (i64.eqz (local.get $n)))
        (br $volta)))
    (if (local.get $neg)
      (then
        (local.set $p (i32.sub (local.get $p) (i32.const 1)))
        (i32.store8 (local.get $p) (i32.const 45))))
    (call $putstr (local.get $p) (i32.sub (i32.const 64) (local.get $p)))
  )
)
