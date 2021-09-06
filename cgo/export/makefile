SRC=${PWD}
WORK=${PWD}/out

CGO_LDFLAGS='"-g" "-O2" "-static"'
GODIR=/opt/share/local/go

cgo: clean
	mkdir -p ${WORK}/b001
	cd ${SRC}
	${GODIR}/pkg/tool/linux_amd64/cgo -debug-gcc -objdir ${WORK}/b001/ -importpath ${SRC}  -- -I  ${WORK}/b001/ -g -O2 -I . ./export.go

# export1
export1: clean
	go build -o libgomodule.so -buildmode=c-shared export1.go
	go build -x -o export1_main export1_main.go
	./export1_main

# export2
libwx.a: clean
	go build -buildmode=c-archive -o libwx.a export2.go

libwx.so: clean
	go build -buildmode=c-shared -o libwx.so export2.go

export2_go_c: libwx.so
	go build -work -o export2_go_c export2_go_c.go
	./export2_go_c

export2_go_c_go: libwx.so
	go build -work -v -x -o export2_go_c_go export2_go_c_go.go
	./export2_go_c_go

clean:
	rm -rf export1_main libgomodule.* export2_go_c export2_go_c_go libwx.*
	rm -rf ${WORK}

.PHONY: init cgo