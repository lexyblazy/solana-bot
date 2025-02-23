build:
	rm -rf ./bin && CGO_ENABLED=1 go build -o bin/app

run: 
	CGO_ENABLED=1 go run solana-bot