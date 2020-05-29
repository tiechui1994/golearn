export LIBRARY_PATH=$(CURDIR)
SRC = $(wildcard *.go)

gomain: $(SRC) libbuffer.a
	go build -o gomain -ldflags "-w" -x $(SRC)

ctest: libbuffer.a
	$(CC) -o ctest test.c -l buffer -l stdc++

libbuffer.a:
	$(CXX) $(CXXFLAGS) -c buffer_c.cpp buffer_cpp.cpp
	$(AR) -r libbuffer.a buffer_c.o buffer_cpp.o

.PHONY : clean
clean:
	-rm -rf ctest gomain *.o *.a *.gch


