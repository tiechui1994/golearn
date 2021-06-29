SRC=${PWD}
WORK=${PWD}/out

CGO_LDFLAGS='"-g" "-O2" "-static"'
GODIR=/opt/share/local/go

cgo: clean
	mkdir -p ${WORK}/b001
	cd ${SRC}
	${GODIR}/pkg/tool/linux_amd64/cgo -debug-gcc -objdir ${WORK}/b001/ -importpath ${SRC}  -- -I  ${WORK}/b001/ -g -O2 -I . ./export.go

build: clean
	go build -x -o libgoadd.so -buildmode=c-shared ./export.go
	go build -race -x -o main ./main.go

clean:
	rm -rf main
	rm -rf ${WORK}

.PHONY: init cgo