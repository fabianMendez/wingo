.PHONY all: build functions

all: build functions

build:
	go build -o main cmd/cli/main.go

functions:
	go build -o netlify/functions/create_notification_setting cmd/functions/create_notification_setting/main.go
	go build -o netlify/functions/confirm_notification_setting cmd/functions/confirm_notification_setting/main.go
	go build -o netlify/functions/get_routes cmd/functions/get_routes/main.go