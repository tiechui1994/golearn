export LIBRARY_PATH=$(CURDIR)
SRC=$(wildcard *.cpp)
OBJ=$(wildcard *.o)

atest: libmid.a
	$(CC) -o test test.c -l mid

libmid.a:
	$(CXX) -c $(SRC)
	$(AR) -r libmid.a $(OBJ)
	@rm -rf *.o

sotest: libmid.so
	$(CC) -o test test.c -l mid

libmid.so:
	$(CXX) -fPIC -shared -o libmid.so $(SRC)


.PHONY: clean
clean:
	-rm -rf *test *.a *.o *.so