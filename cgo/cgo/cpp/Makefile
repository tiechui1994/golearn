export LIBRARY_PATH=$(CURDIR)
SRC = $(wildcard *.go)

gotest: $(SRC) libbuffer.a
	go build -o gotest -ldflags "-w" -x $(SRC)

ctest: libbuffer.a
	$(CC) -o ctest test.c -lbuffer -lstdc++

libbuffer.a:
	$(CXX) $(CXXFLAGS) -c *.cpp
	$(AR) -r libbuffer.a *.o
	@rm -r *.o

.PHONY : clean
clean:
	-rm -rf ctest gotest *.o *.a *.gch


