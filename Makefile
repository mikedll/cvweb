
all: bin/cli bin/web_server bin/debug_web_server

bin/cli: $(wildcard pkg/*.go) cmd/cli/main.go
	go build -o bin/cli ./cmd/cli

bin/web_server: $(wildcard pkg/*.go) $(wildcard cmd/web_server/*.go)
	go build -o bin/web_server ./cmd/web_server

bin/debug_web_server: $(wildcard pkg/*.go) $(wildcard cmd/web_server/*.go)
	go build -o bin/debug_web_server -tags matprofile ./cmd/web_server

debug_web_server: bin/debug_web_server

clean:
	rm bin/*
