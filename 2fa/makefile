all: client server

client:
	go build -o client -ldflags='-w -s' client.go util.go

server:
	go build -o server -ldflags='-w -s' server.go util.go


.PHONY:clean

clean:
	@rm -rf client server