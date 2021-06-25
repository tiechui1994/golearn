.PHONY:env
env:
	@echo PATH=$(PATH)
	@echo C_INCLUDE_PATH=$(C_INCLUDE_PATH)
	@echo CPLUS_INCLUDE_PATH=$(CPLUS_INCLUDE_PATH)
	@echo LD_LIBRARY_PATH=$(LD_LIBRARY_PATH)
	@echo LIBRARY_PATH=$(LIBRARY_PATH)
	@echo CFLAGS=$(CFLAGS)
	@echo CXXFLAGS=$(CXXFLAGS)
	@echo LDFLAGS=$(LDFLAGS)
	@echo CC=$(CC)
	@echo CXX=$(CXX)
	@echo AR=$(AR)
	@echo CURDIR=$(CURDIR)
	@echo SHELL=$(SHELL)
	@echo MAKEFLAGS=$(MAKEFLAGS)
	@echo TEST=$(TEST)

libwx.a: clean
	go build -buildmode=c-archive -o libwx.a export1.go

libwx.so: clean
	go build -buildmode=c-shared -o libwx.so export1.go

main: libwx.so
	go build -o main export2.go

clean:
	rm -rf libwx.*
	rm -rf main

all: libwx.a libwx.so