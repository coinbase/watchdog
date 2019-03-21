SHA := `git rev-parse HEAD`
TIME := `date -u '+%Y-%m-%d_%I:%M:%S%p'`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.BuildSha=$(SHA) -X=main.BuildTime=$(TIME)"

.PHONY: install

test:
	bash ./scripts/test.sh

gosec:
	# move a gosec to a separate target because it cannot run with modules
	# https://github.com/securego/gosec/issues/234
	GO111MODULE=off go get github.com/securego/gosec/cmd/gosec/...
	gosec -exclude=G106 ./...

install:
	go install $(LDFLAGS)
