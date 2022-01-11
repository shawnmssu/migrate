TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

test:
	go test $(TEST) -timeout=30s -parallel=32

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w -s $(GOFMT_FILES)

clean:
	rm -rf bin/*


install:
	go build go build -o bin/ ./...
	chmod +x bin/migrate

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate

mac-arm:
	GOOS=darwin GOARCH=arm64 go build -o bin/ ./...
	chmod +x bin/migrate

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate
