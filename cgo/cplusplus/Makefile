export LIBRARY_PATH=$(CURDIR)
SRC=$(wildcard *.cpp)
OBJ=$(wildcard *.o)

atest:libxxx_a
	$(CC) -o test test.c -lxxx

libxxx_a:
	$(CXX) -c $(SRC)
	$(AR) -r libxxx.a $(OBJ)
	@rm -rf $(OBJ)

sotest:libxxx_so
	$(CC) -o test test.c -lxxx

libxxx_so:
	$(CXX) -fPIC -shared -o libxxx.so $(SRC)

.PHONY: clean
clean:
	-rm -rf *test *.a *.o *.so