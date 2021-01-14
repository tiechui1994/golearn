all: client server test

client:
	@rm -rf client
	go build -o client -ldflags='-w -s' client.go util.go

server:
	@rm -rf server
	go build -o server -ldflags='-w -s' server.go util.go

test:
	@rm -rf test
	go build -o test -ldflags='-w -s' main.go util.go


.PHONY:clean

clean:
	@rm -rf client server test
