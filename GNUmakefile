export VERSION=0.1.0
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

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

all: clean mac mac-arm windows linux

install:clean
	mkdir -p bin
	go build -o bin/ ./...
	chmod +x bin/migrate
	cp bin/migrate /usr/local/bin/migrate

mac:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate
	cd bin/ && tar czvf ./migrate_${VERSION}_darwin-amd64.tgz ./migrate
	rm -rf ./bin/migrate

mac-arm:
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -o bin/ ./...
	chmod +x bin/migrate
	cd bin/ && tar czvf ./migrate_${VERSION}_darwin-arm64.tgz ./migrate
	rm -rf ./bin/migrate

windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate.exe
	cd bin/ && tar czvf ./migrate_${VERSION}_windows-amd64.tgz ./migrate.exe
	rm -rf ./bin/migrate.exe

linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/ ./...
	chmod +x bin/migrate
	cd bin/ && tar czvf ./migrate_${VERSION}_linux-amd64.tgz ./migrate
	rm -rf ./bin/migrate
