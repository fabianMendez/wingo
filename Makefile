.PHONY all: build functions

all: build functions

build:
	cd cmd && go build -o ../main && cd -

functions:
	cd cmd && go build -o ../netlify/functions/create_notification_setting -tags 'create_notification_setting' && cd -