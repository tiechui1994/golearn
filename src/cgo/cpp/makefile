SRC = $(wildcard *.go)
test:$(SRC)
#	g++ -c -o buffer.o buffer_c.c
#	ar rcs libbuffer.a buffer.o
	go build -o test -x -ldflags "-w -s" ${SRC}



