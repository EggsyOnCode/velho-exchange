build:
	go build -o ./bin/vleho

run: build
	./bin/vleho

test: 
	go test -v ./...