build:
	go build -o ./bin/cs

run: build
	./bin/cs