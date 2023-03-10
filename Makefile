build:
	go build -o ./lostsonstv

run: build
	./lostsonstv

test:
	go test ./...