
all: bin/cli

bin/cli: $(wildcard pkg/*.go) cmd/cli/main.go
	go build -o bin/cli cmd/cli/main.go

clean:
	rm bin/*
