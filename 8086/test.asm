mov si, bx
mov dh, al
mov cl, 12
mov ch, -12
mov cx, 12
mov cx, -12
mov dx, 3948
mov dx, -3948
mov al, [bx + si]
mov bx, [bp + di]
mov dx, [bp]
mov ah, [bx + si + 4]
mov al, [bx + si + 4999]
mov [bx + di], cx
mov [bp + si], cl
mov [bp], ch
mov ax, [bx + di - 37]
mov [si - 300], cx
mov dx, [bx - 32]
mov [bp + di], byte 7
mov [di + 901], word 347
mov bp, [5]
mov bx, [3458]
mov ax, [2555]
mov ax, [16]
mov [2554], ax
mov [15], ax

push word [bp + si]
push word [3000]
push word [bx + di - 30]
push cx
push ax
push dx
push cs

pop word [bp + si]
pop word [3]
pop word [bx + di - 3000]
pop sp
pop di
pop si
pop ds

add cx, 0x23
add cx, [bp]
add dx, [bx + si]
add [bp + di + 5000], ah
add [bx], al
add sp, 392
add si, 5
add ax, 1000
add ah, 30
add al, 9
add cx, bx
add ch, al

inc ax
inc cx
inc dh
inc al
inc ah
inc sp
inc di
inc byte [bp + 1002]
inc word [bx + 39]
inc byte [bx + si + 5]
inc word [bp + di - 10044]
inc word [9349]
inc byte [bp]

sub cx, [bp]
sub dx, [bx + si]
sub [bp + di + 5000], ah
sub [bx], al
sub sp, 392
sub si, 5
sub ax, 1000
sub ah, 30
sub al, 9
sub cx, bx
sub ch, al

xchg ax, [bp - 1000]
xchg [bx + 50], bp
xchg ax, ax
xchg ax, dx
xchg ax, sp
xchg ax, si
xchg ax, di
xchg cx, dx
xchg si, cx
xchg cl, ah

in al, 200
in al, dx
in ax, dx

out 44, ax
out dx, al
