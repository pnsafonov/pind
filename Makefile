.PHONY: build

build: pind

pind:
	go build -v -o ./pind github.com/pnsafonov/pind

test:
	go test -count=1 ./...
