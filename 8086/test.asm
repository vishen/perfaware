bits 16

; sub sp, 392
; sub si, 5
; sub ah, 30
;xchg [bx + 50], bp
;xchg [bx - 1000], bp
xchg [bp - 1000], ax
xchg ax, [bp - 1000]
;xchg cx, dx
;xchg si, cx
;xchg cl, ah
