util= baidu.go dns.go init.go speechify.go speechtext.go

clean:
	rm -rf lightyearvpn
	rm -rf teambition
	rm -rf drive
	rm -rf idea
	rm -rf gitee
	rm -rf login

lightyearvpn: clean
	go build -o lightyearvpn -v -ldflags '-w' $(util)  lightyearvpn.go

teambition: clean
	go build -o teambition -v -ldflags '-w' $(util)  teambition.go

drive: clean
	go build -o drive -v -ldflags '-w' $(util)  drive.go

idea: clean
	go build -o idea -v -ldflags '-w' $(util)  idea.go

gitee: clean
	go build -o gitee -v -ldflags '-w' $(util)  gitee.go

login: clean
	go build -o login -v -ldflags '-w' $(util)  login.go

all: lightyearvpn teambition drive idea gitee login

