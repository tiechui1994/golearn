export LIBRARY_PATH=$(CURDIR)

atest: libmid.a
	$(CC) -o test test.c -l mid

libmid.a:
	$(CXX) -c mid.cpp cxx.cpp
	$(AR) -r libmid.a mid.o cxx.o

sotest: libmid.so
	$(CC) -o test test.c -l mid

libmid.so:
	$(CXX) -fPIC -shared -o libmid.so cxx.cpp mid.cpp


.PHYON: clean
clean:
	-rm -rf *test *.a *.o *.so