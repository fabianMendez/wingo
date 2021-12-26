.PHONY all: build functions

all: build functions

build:
	go build -o main cmd/main.go

functions:
	go build -o netlify/functions/create_subscription netlify/functions/create_subscription/main.go
	go build -o netlify/functions/confirm_subscription netlify/functions/confirm_subscription/main.go
	go build -o netlify/functions/get_routes netlify/functions/get_routes/main.go