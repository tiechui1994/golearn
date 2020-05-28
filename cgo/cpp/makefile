export LIBRARY_PATH=$(CURDIR)
SRC = $(wildcard *.go)

main: $(SRC) libbuffer.a
	go build -o main -ldflags "-w -z" -x $(SRC)

test: libbuffer.a
	$(CC) -o test test.c -l buffer

libbuffer.a:
	$(CXX) -c buffer_c.cpp buffer_cpp.cpp
	$(AR) -r libbuffer.a buffer_c.o buffer_cpp.o

.PHYON : clean
clean:
	-rm -rf test main *.o *.a *.gch


