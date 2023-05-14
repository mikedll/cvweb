
all: bin/cli bin/web_server

bin/cli: $(wildcard pkg/*.go) cmd/cli/main.go
	go build -o bin/cli cmd/cli/main.go

bin/web_server: $(wildcard pkg/*.go) cmd/web_server/main.go
	go build -o bin/web_server cmd/web_server/main.go

bin/debug_web_server: $(wildcard pkg/*.go) cmd/web_server/main.go
	go build -o bin/debug_web_server -tags matprofile cmd/web_server/main.go

debug_web_server: bin/debug_web_server

clean:
	rm bin/*
