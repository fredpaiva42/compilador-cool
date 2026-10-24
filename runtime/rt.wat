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
;; Objetos: todo valor COOL e um endereco i32 (0 = void). Cabecalho de 8 bytes:
;; +0 tag da classe, +4 tamanho total. Int: +8 valor i64 (16 bytes).
;; Bool: +8 valor i32 (12 bytes). String: +8 bytes (12 + tamanho).
;; Objeto de usuario: a partir de +8, um campo de 8 bytes por atributo.
;; Tags: Object=1, IO=2, Int=3, String=4, Bool=5; as do programa vao de 6
;; em ordem alfabetica (o emissor calcula em classTags).

;; >>> RUNTIME (o emissor copia daqui)
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

  ;; obj_new(tag, size): aloca, zera e carimba tag e tamanho
  (func $obj_new (param $tag i32) (param $size i32) (result i32)
    (local $p i32)
    (local.set $p
      (call $alloc
        (local.get $size)))
    (memory.fill
      (local.get $p)
      (i32.const 0)
      (local.get $size))
    (i32.store
      (local.get $p)
      (local.get $tag))
    (i32.store offset=4
      (local.get $p)
      (local.get $size))
    (local.get $p))

  ;; new_int(v): encaixota um i64 num objeto Int
  (func $new_int (param $v i64) (result i32)
    (local $p i32)
    (local.set $p
      (call $obj_new
        (i32.const 3)
        (i32.const 16)))
    (i64.store offset=8
      (local.get $p)
      (local.get $v))
    (local.get $p))

  ;; new_bool(v): encaixota um i32 num objeto Bool
  (func $new_bool (param $v i32) (result i32)
    (local $p i32)
    (local.set $p
      (call $obj_new
        (i32.const 5)
        (i32.const 12)))
    (i32.store offset=8
      (local.get $p)
      (local.get $v))
    (local.get $p))

  ;; new_str_empty: String vazia (o passo 6 poe os literais aqui)
  (func $new_str_empty (result i32)
    (call $obj_new
      (i32.const 4)
      (i32.const 12)))

  ;; copy_obj: copia rasa, que o Object.copy usa no passo 5
  (func $copy_obj (param $src i32) (result i32)
    (local $size i32)
    (local $dst i32)
    (local.set $size
      (i32.load offset=4
        (local.get $src)))
    (local.set $dst
      (call $alloc
        (local.get $size)))
    (memory.copy
      (local.get $dst)
      (local.get $src)
      (local.get $size))
    (local.get $dst))

  ;; abort: o Object.abort usa (sai com codigo 1)
  (func $abort
    (call $proc_exit
      (i32.const 1)))

  ;; str_concat(a, b): junta os bytes num objeto novo
  (func $str_concat (param $a i32) (param $b i32) (result i32)
    (local $la i32)
    (local $lb i32)
    (local $p i32)
    (local.set $la
      (i32.load offset=4
        (local.get $a)))
    (local.set $lb
      (i32.load offset=4
        (local.get $b)))
    (local.set $p
      (call $obj_new
        (i32.const 4)
        (i32.add
          (i32.const 12)
          (i32.add
            (local.get $la)
            (local.get $lb)))))
    (i32.store offset=4
      (local.get $p)
      (i32.add
        (local.get $la)
        (local.get $lb)))
    (memory.copy
      (i32.add
        (local.get $p)
        (i32.const 8))
      (i32.add
        (local.get $a)
        (i32.const 8))
      (local.get $la))
    (memory.copy
      (i32.add
        (i32.add
          (local.get $p)
          (i32.const 8))
        (local.get $la))
      (i32.add
        (local.get $b)
        (i32.const 8))
      (local.get $lb))
    (local.get $p))

  ;; str_substr(s, i, l): fatia com indices 0-based; fora do alcance aborta
  (func $str_substr (param $s i32) (param $i i64) (param $l i64) (result i32)
    (local $len i32)
    (local $p i32)
    (local.set $len
      (i32.load offset=4
        (local.get $s)))
    (if
      (i32.or
        (i64.lt_s
          (local.get $i)
          (i64.const 0))
        (i32.or
          (i64.lt_s
            (local.get $l)
            (i64.const 0))
          (i64.gt_u
            (i64.add
              (local.get $i)
              (local.get $l))
            (i64.extend_i32_u
              (local.get $len)))))
      (then
        (call $abort)))
    (local.set $p
      (call $obj_new
        (i32.const 4)
        (i32.add
          (i32.const 12)
          (i32.wrap_i64
            (local.get $l)))))
    (i32.store offset=4
      (local.get $p)
      (i32.wrap_i64
        (local.get $l)))
    (memory.copy
      (i32.add
        (local.get $p)
        (i32.const 8))
      (i32.add
        (i32.add
          (local.get $s)
          (i32.const 8))
        (i32.wrap_i64
          (local.get $i)))
      (i32.wrap_i64
        (local.get $l)))
    (local.get $p))

  ;; str_eq(a, b): 1 se mesmo tamanho e mesmos bytes, 0 se nao
  (func $str_eq (param $a i32) (param $b i32) (result i32)
    (local $la i32)
    (local $lb i32)
    (local $i i32)
    (if
      (i32.eq
        (local.get $a)
        (local.get $b))
      (then
        (return
          (i32.const 1))))
    (local.set $la
      (i32.load offset=4
        (local.get $a)))
    (local.set $lb
      (i32.load offset=4
        (local.get $b)))
    (if
      (i32.ne
        (local.get $la)
        (local.get $lb))
      (then
        (return
          (i32.const 0))))
    (local.set $i
      (i32.const 0))
    (block $fim
      (loop $volta
        (br_if $fim
          (i32.ge_u
            (local.get $i)
            (local.get $la)))
        (if
          (i32.ne
            (i32.load8_u
              (i32.add
                (i32.add
                  (local.get $a)
                  (i32.const 8))
                (local.get $i)))
            (i32.load8_u
              (i32.add
                (i32.add
                  (local.get $b)
                  (i32.const 8))
                (local.get $i))))
          (then
            (return
              (i32.const 0))))
        (local.set $i
          (i32.add
            (local.get $i)
            (i32.const 1)))
        (br $volta)))
    (i32.const 1))

  ;; io_out_string(s): imprime os bytes do objeto
  (func $io_out_string (param $s i32)
    (call $putstr
      (i32.add
        (local.get $s)
        (i32.const 8))
      (i32.load offset=4
        (local.get $s))))

  ;; io_out_int(v): imprime o i64 desembrulhado
  (func $io_out_int (param $v i64)
    (call $putint
      (local.get $v)))

  ;; io_in_string: le a linha para um objeto novo (o manual exclui o \n)
  (func $io_in_string (result i32)
    (local $p i32)
    (local $n i32)
    (local.set $p
      (call $obj_new
        (i32.const 4)
        (i32.const 1036)))
    (local.set $n
      (call $read_line
        (i32.add
          (local.get $p)
          (i32.const 8))
        (i32.const 1024)))
    (i32.store offset=4
      (local.get $p)
      (local.get $n))
    (local.get $p))

  ;; io_in_int: le o inteiro da entrada (o runtime descarta o resto da linha)
  (func $io_in_int (result i64)
    (call $read_int))
  ;; <<< RUNTIME (ate aqui)

  ;; mensagens e literais
  (data (i32.const 2048) "\n")
  (data (i32.const 2056) "alloc ok\n")
  (data (i32.const 2072) "alloc FALHOU\n")
  (data (i32.const 2088) "int: ")
  (data (i32.const 2096) "str: ")
  (data (i32.const 2104) "linha: ")
  (data (i32.const 2112) "soma: ")
  (data (i32.const 2120) "neg: ")
  (data (i32.const 2128) "fim\n")
  (data (i32.const 2136) "ola, compilador")

  ;; bateria do runtime: exercita alocação, inteiros, strings e entrada.
    ;; confira com testar.sh (comparação contra bateria.esperado)
    (func (export "_start")
      (local $p1 i32)
      (local $p2 i32)
      (local $p3 i32)
      (local $ok i32)
      (local $n i32)

      ;; 1. alocador: 8, 9 e 1 byte devem devolver pulos de 8, 16 e 8
      (local.set $p1 (call $alloc (i32.const 8)))
      (local.set $p2 (call $alloc (i32.const 9)))
      (local.set $p3 (call $alloc (i32.const 1)))
      (local.set $ok
        (i32.and
          (i32.eq (i32.sub (local.get $p2) (local.get $p1)) (i32.const 8))
          (i32.eq (i32.sub (local.get $p3) (local.get $p2)) (i32.const 16))))
      (if (local.get $ok)
        (then (call $putstr (i32.const 2056) (i32.const 9)))
        (else (call $putstr (i32.const 2072) (i32.const 13))))

      ;; 2. inteiros: zero, negativo de varios digitos e o maior i64
      (call $putstr (i32.const 2088) (i32.const 5))
      (call $putint (i64.const 0))
      (call $putstr (i32.const 2048) (i32.const 1))
      (call $putstr (i32.const 2088) (i32.const 5))
      (call $putint (i64.const -1234567890))
      (call $putstr (i32.const 2048) (i32.const 1))
      (call $putstr (i32.const 2088) (i32.const 5))
      (call $putint (i64.const 9223372036854775807))
      (call $putstr (i32.const 2048) (i32.const 1))

      ;; 3. string literal dos dados estaticos
      (call $putstr (i32.const 2096) (i32.const 5))
      (call $putstr (i32.const 2136) (i32.const 15))
      (call $putstr (i32.const 2048) (i32.const 1))

      ;; 4. entrada: uma linha, dois numeros (o primeiro descarta o resto
      ;; da linha) e um negativo
      (local.set $n (call $read_line (i32.const 1100) (i32.const 200)))
      (call $putstr (i32.const 2104) (i32.const 7))
      (call $putstr (i32.const 1100) (local.get $n))
      (call $putstr (i32.const 2048) (i32.const 1))
      (call $putstr (i32.const 2112) (i32.const 6))
      (call $putint (i64.add (call $read_int) (call $read_int)))
      (call $putstr (i32.const 2048) (i32.const 1))
      (call $putstr (i32.const 2120) (i32.const 5))
      (call $putint (call $read_int))
      (call $putstr (i32.const 2048) (i32.const 1))

      ;; fim
      (call $putstr (i32.const 2128) (i32.const 4))
      (call $proc_exit (i32.const 0)))
)
