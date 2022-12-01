export LIBRARY_PATH=$(CURDIR)
CPLUS=$(wildcard *.cpp)
C=$(wildcard *.c)

test1: static
	$(CC) -o test1 $(C) -L. -lfunc -static

static: clean
	$(CXX) -c -std=c++11 $(CPLUS)
	$(AR) -r libfunc.a *.o
	@rm -rf *.o

test2: dynamic
	$(CC) -o test2 $(C) -L. -lfunc -lstdc++ -Wl,-rpath=. 

dynamic: clean
	$(CXX) -o libfunc.so $(CPLUS) -fPIC -shared -std=c++11 

.PHONY: clean
clean:
	-rm -rf test1 test2 *.a *.o *.so