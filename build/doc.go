package build

/*
	go build -ldflags \
		"-X github.com/atsu/goat/build.Version=${VERSION} \
		-X github.com/atsu/goat/build.CommitHash=${COMMIT_HASH} \
		-X github.com/atsu/goat/build.Date=${DATE} \
		-X github.com/atsu/goat/build.Agent=${AGENT}"
*/
