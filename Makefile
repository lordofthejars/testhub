version ?= latest

build:
	go build -o testhub

cross:
	docker run -it --rm -v "$$PWD":/go/src/github.com/lordofthejars/testhub -w /go/src/github.com/lordofthejars/testhub -e "version=${version}" lordofthejars/goreleaser:1.0 crossbuild.sh