SRC=$(wildcard *.c)

# warning: curent lib can generate shared lib, not static lib

main: libheader.so
	go build -o main main.go

libheader.so:
	$(CC) -fPIC -shared -o libheader.so  $(SRC)

.PYTHON: clean
clean:
	-rm -rf *.a *.o *.so main