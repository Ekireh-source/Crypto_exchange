build: 
	go build -o bin/exchange

run: build
	./bin/exchange

migrate: build
	./bin/exchange migrate

test:
	go test -v ./...

swagger:
	swag init -g main.go -o docs