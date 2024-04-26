SRC=$(wildcard *.c)

# warning: current lib can generate shared lib, not static lib

main: libheader.so
	go build -o main main.go
	./main

libheader.so: clean
	$(CC) -fPIC -shared -o libheader.so  $(SRC)

clean:
	-rm -rf *.a *.o *.so main

.PYTHON: clean