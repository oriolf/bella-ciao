COMMIT=$(shell git log --pretty=format:'%h' -n 1)
build:
	go build -ldflags "-X main.commitHash=${COMMIT}"
