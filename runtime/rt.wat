;; Runtime COOL -> WASM. Convenções de memória:
;; 0..7 iovec de escrita {ptr, len}
;; 8..11 nwritten
;; 16..23 iovec de leitura {ptr, len}
;; 24..27 nread
;; 64..1087 buffer de leitura (1024 bytes)
;; 2048 strings literais (segmentos de dados)
;; 65536.. heap (objetos), cresce para cima
(module
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))

  (memory (export "memory") 16)

  ;; próximo endereço livre do heap
  (global $heap (mut i32) (i32.const 65536))

  (data
    (i32.const 2048) "alloc ok\n")
  (data (i32.const 2080) "alloc FALHOU\n")

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
    (local $p1 i32)
    (local $p2 i32)
    (local $p3 i32)
    (local $ok i32)
    (local.set $p1 (call $alloc (i32.const 8)))
    (local.set $p2 (call $alloc (i32.const 9)))
    (local.set $p3 (call $alloc (i32.const 1)))
    ;; ok = (p2 - p1 == 8) && (p3 - p2 == 16)
    (local.set $ok
      (i32.and
        (i32.eq (i32.sub (local.get $p2) (local.get $p1)) (i32.const 8))
        (i32.eq (i32.sub (local.get $p3) (local.get $p2)) (i32.const 16))))
    (if (local.get $ok)
      (then (call $putstr
        (i32.const 2048) (i32.const 9)))
      (else (call $putstr (i32.const 2080) (i32.const 13))))
    (call $proc_exit (i32.const 0)))

  ;; alloc(size): devolve o próximo endereço livre e avança o heap
  ;; O tamanho e arrendondado para múltiplo de 8: (size + 7) & -8
  (func $alloc (param $size i32) (result i32)
    (local $p i32)
    (local.set $p (global.get $heap))
    (global.set $heap
      (i32.add (local.get $p)
        (i32.and (i32.add (local.get $size) (i32.const 7 )) (i32.const -8))))
    (local.get $p)))
