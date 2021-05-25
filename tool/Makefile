util= baidu.go dns.go init.go

clean:
	rm -rf gitee
	rm -rf speechtext
	rm -rf speechify
	rm -rf login
	rm -rf lightyearvpn
	rm -rf drive
	rm -rf teambition
	rm -rf idea

gitee: clean
	go build -o gitee -v -ldflags '-w' $(util)  gitee.go

speechtext: clean
	go build -o speechtext -v -ldflags '-w' $(util)  speechtext.go

speechify: clean
	go build -o speechify -v -ldflags '-w' $(util)  speechify.go

login: clean
	go build -o login -v -ldflags '-w' $(util)  login.go

lightyearvpn: clean
	go build -o lightyearvpn -v -ldflags '-w' $(util)  lightyearvpn.go

drive: clean
	go build -o drive -v -ldflags '-w' $(util)  drive.go

teambition: clean
	go build -o teambition -v -ldflags '-w' $(util)  teambition.go

idea: clean
	go build -o idea -v -ldflags '-w' $(util)  idea.go

all: gitee speechtext speechify login lightyearvpn drive teambition idea

