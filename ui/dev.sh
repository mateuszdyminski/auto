#!/bin/bash

build() {
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ui -a -tags netgo .
}

buildDocker() {
	docker build -t mateuszdyminski/auto-ui:latest .
}

pushDocker() {
	docker push mateuszdyminski/auto-ui
}

CMD="$1"
SUBCMD="$2"

shift
case "$CMD" in
	build)
		build
	;;
	docker-build)
		buildDocker
	;;
    docker-push)
		pushDocker
	;;
	all)
		build
		buildDocker
		pushDocker
	;;
	*)
		echo 'Choose one of following args: {build, docker-build, docker-push, all}'
	;;
esac