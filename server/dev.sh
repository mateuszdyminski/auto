#!/bin/bash

build() {
	GITREPO='github.com/mateuszdyminski/auto/server'
	APP_VERSION='0.1'
	GIT_VERSION=$(git describe --always)
	LAST_COMMIT_USER="$(tr -d '[:space:]' <<<"$(git log -1 --format=%cn)<$(git log -1 --format=%ce)>")"
	LAST_COMMIT_HASH=$(git log -1 --format=%H)
	LAST_COMMIT_TIME=$(git log -1 --format=%cd --date=format:'%Y-%m-%d_%H:%M:%S')

	LDFLAGS="-s -w -X $GITREPO/pkg/version.APP_VERSION=$APP_VERSION -X $GITREPO/pkg/version.GIT_VERSION=$GIT_VERSION -X $GITREPO/pkg/version.LAST_COMMIT_TIME=$LAST_COMMIT_TIME -X $GITREPO/pkg/version.LAST_COMMIT_HASH=$LAST_COMMIT_HASH -X $GITREPO/pkg/version.LAST_COMMIT_USER=$LAST_COMMIT_USER -X $GITREPO/pkg/version.BUILD_TIME=$(date -u +%Y-%m-%d_%H:%M:%S)"
	
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o build/server -a -tags netgo ./cmd/server
}

buildDocker() {
	docker build -t mateuszdyminski/auto-server:latest .
}

pushDocker() {
	docker push mateuszdyminski/auto-server
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
