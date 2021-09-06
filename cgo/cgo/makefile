inline: clean
	go build -o inline -ldflags '-w -s' cgo_inline_c.go
	./inline

libmath.a: clean
	$(CC) -c -fPIC math/math.c
	$(AR) -r libmath.a math.o

library: libmath.a
	go build -o library -ldflags '-w -s --extldflags "-static -fPIC"' cgo_inline_clib.go
	./library

outline: clean
	go build -o outline -ldflags '-w -s' cgo_outline_c.go
	./outline

clean:
	@rm -rf inline library libmath.a *.o outline

.PHONY:clean