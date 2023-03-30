bits 16

; mov al, cs:[bx + si]
; mov bx, cs:[bp]

and cs:[bp + si], bx
; rep movsb
; mov al, ds:[bx + si]
