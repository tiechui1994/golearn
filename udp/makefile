udp: clean
	go build -o server -ldflags='-X main.server=SimpleCastServer' server.go
	go build -o client -ldflags='-X main.client=SimpleCastClient' client.go

broadcast: clean
	go build -o server -ldflags='-X main.server=BroadCastServer' server.go
	go build -o client -ldflags='-X main.client=BroadCastClient' client.go

multi: clean
	go build -o server -ldflags='-X main.server=MultiCastServer' server.go
	go build -o client -ldflags='-X main.client=MultiCastClient' client.go

clean:
	@rm -rf server client
