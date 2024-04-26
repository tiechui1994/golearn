export LIBRARY_PATH=$(CURDIR)
SRC = $(wildcard *.go)

gotest: $(SRC) static
	go build -o gotest -ldflags "-w" -x $(SRC)

ctest: static
	$(CC) -o ctest main.c -lbuffer -lstdc++

static:
	$(CXX) $(CXXFLAGS) -c *.cpp
	$(AR) -r libbuffer.a *.o
	@rm -r *.o

.PHONY : clean
clean:
	-rm -rf ctest gotest *.o *.a *.gch


