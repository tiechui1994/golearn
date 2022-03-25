crash: lib
	go build -o crash -gcflags=all="-N -l" crash.go main_crash.go

print: lib
	go build -o print -gcflags=all="-N -l" print.go main_print.go

lib:
	gcc -I. -g3 -O0 -c test.cpp
	ar cr libtest.a test.o

clean:
	@rm -rf *.o *.a crash print