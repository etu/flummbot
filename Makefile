GOPATH=`pwd`/vendor

build:
	env GOPATH=$(GOPATH) go build -o flummbot
