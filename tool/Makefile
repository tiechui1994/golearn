util= baidu.go dns.go init.go

clean:
	rm -rf lightyearvpn
	rm -rf speechtext
	rm -rf speechify
	rm -rf teambition
	rm -rf drive
	rm -rf idea
	rm -rf gitee

lightyearvpn: clean
	go build -o lightyearvpn -v -ldflags '-w' $(util)  lightyearvpn.go

speechtext: clean
	go build -o speechtext -v -ldflags '-w' $(util)  speechtext.go

speechify: clean
	go build -o speechify -v -ldflags '-w' $(util)  speechify.go

teambition: clean
	go build -o teambition -v -ldflags '-w' $(util)  teambition.go

drive: clean
	go build -o drive -v -ldflags '-w' $(util)  drive.go

idea: clean
	go build -o idea -v -ldflags '-w' $(util)  idea.go

gitee: clean
	go build -o gitee -v -ldflags '-w' $(util)  gitee.go

all: lightyearvpn speechtext speechify teambition drive idea gitee

