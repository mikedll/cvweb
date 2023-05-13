
all: bin/cli bin/web_server

bin/cli: $(wildcard pkg/*.go) cmd/cli/main.go
	go build -o bin/cli cmd/cli/main.go

bin/web_server: $(wildcard pkg/*.go) cmd/web_server/main.go
	go build -o bin/web_server cmd/web_server/main.go

clean:
	rm bin/*
