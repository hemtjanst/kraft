
SOURCES := $(shell find . -name "*.go")

all: kraft_linux_amd64 kraft_linux_arm64 kraft_linux_arm7

go.sum: go.mod
	go mod tidy

test: $(SOURCES) go.sum
	go test ./...

kraft_linux_amd64: $(SOURCES) go.sum
	env GOOS=linux GOARCH=amd64 go build -o kraft_linux_amd64 -ldflags '-s -w' hemtjan.st/kraft

kraft_linux_arm64: $(SOURCES) go.sum
	env CC=arm-none-eabi-gcc CGO_ENABLED=0  GOOS=linux GOARCH=arm64 go build -buildmode=exe -o kraft_linux_arm64 -ldflags '-extldflags "-fno-PIC static" -s -w' -tags 'osusergo netgo static_build' hemtjan.st/kraft

kraft_linux_arm7: $(SOURCES) go.sum
	env CC=arm-none-eabi-gcc CGO_ENABLED=0  GOOS=linux GOARCH=arm GOARM=7 go build -buildmode=exe -o kraft_linux_arm7 -ldflags '-extldflags "-fno-PIC static" -s -w' -tags 'osusergo netgo static_build' hemtjan.st/kraft

