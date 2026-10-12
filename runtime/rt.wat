;; Runtime COOL -> WASM. Convenções de memória:
;; 0..7 iovec de escrita {ptr, len}
;; 8..11 nwritten
;; 16..23 iovec de leitura {ptr, len}
;; 24..27 nread
;; 64..1087 buffer de leitura (1024 bytes)
;; 2048 strings literais (segmentos de dados)
;; 65536.. heap(objetos), cresce para cima
(module
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))

  (memory (export "memory") 16)

  (data
    (i32.const 2048) "runtime ok\n")

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
    (call $putstr
      (i32.const 2048)
      (i32.const 11))
    (call $proc_exit
      (i32.const 0))))
