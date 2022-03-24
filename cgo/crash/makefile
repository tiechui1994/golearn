build:
	gcc -I. -g3 -O0 -c test.cpp
	ar cr libtest.a test.o
	go build -o test -gcflags=all="-N -l" crash.go main.go

clean:
	@rm -rf *.o *.a test