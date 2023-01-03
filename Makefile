build:
ifeq ($(OS),Windows_NT)
	go build -o build/ins-bsc-tx-tool.exe main.go
else
	go build -o build/ins-bsc-tx-tool main.go
endif

install:
ifeq ($(OS),Windows_NT)
	go install main.go
else
	go install main.go
endif

.PHONY: build install
