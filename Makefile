APP=stingray

.PHONY: build run test clean

build:
	go build -o $(APP) .

run:
	go run .

test:
	go test ./...

clean:
	rm -f $(APP) 