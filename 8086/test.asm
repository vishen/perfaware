bits 16

; shl ah, 1
;shl ah, 5
; shl ah, cl
; test bx, cx
; test bl, 20
; xor al, 93
; xor bx, 93
; xor byte [bp - 39], 239
; xor [bx + di + 1000], dx
; xor byte [bp - 39], 239
rep movsb
