CPLUS=$(wildcard *.cpp)
GO=$(wildcard *.go)
C=$(wildcard *.c)

gomain: static
	go build -o gomain $(GO)

cmain: dynamic
	$(CC) -o cmain main.c -L. -ldate -lstdc++ -Wl,-rpath=.

static: clean
	$(CXX) -c -std=c++11 -D __cplusplus_combine $(CPLUS)
	$(AR) -r libdate.a *.o
	@rm -rf *.o

dynamic: clean
	$(CXX) -fPIC -shared -o libdate.so -std=c++11 -D __cplusplus_inherit $(CPLUS)

.PHONY : clean
clean:
	@rm -rf cmain gomain  *.o *.a *.so *.gch