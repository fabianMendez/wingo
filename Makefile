.PHONY all: run functions

all: run functions

run:
	mkdir -p build
	go build -o build/run cmd/*.go

functions:
	mkdir -p build/functions
	go build -o build/functions/create_subscription netlify/functions/create_subscription/main.go
	go build -o build/functions/confirm_subscription netlify/functions/confirm_subscription/main.go