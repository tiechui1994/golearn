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
	go build -o lightyearvpn -v -ldflags '-w -s -X main.AppName=lightyearvpn' $(util) lightyearvpn.go

speechtext: clean
	go build -o speechtext -v -ldflags '-w -s -X main.AppName=speechtext' $(util) speechtext.go

speechify: clean
	go build -o speechify -v -ldflags '-w -s -X main.AppName=speechify' $(util) speechify.go

teambition: clean
	go build -o teambition -v -ldflags '-w -s -X main.AppName=teambition' $(util) teambition.go

drive: clean
	go build -o drive -v -ldflags '-w -s -X main.AppName=drive' $(util) drive.go

idea: clean
	go build -o idea -v -ldflags '-w -s -X main.AppName=idea' $(util) idea.go

gitee: clean
	go build -o gitee -v -ldflags '-w -s -X main.AppName=gitee' $(util) gitee.go

all: lightyearvpn speechtext speechify teambition drive idea gitee

