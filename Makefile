GOFMT = gofmt -s
FILES = $(shell find . -name "*.go")
VETPACKAGES = $(shell go list ./... | grep -v /test/)

ci:
	make vet
	make fmt-check
	make test
	make test-race
	make cover

.PHONY: vet
vet:
	@go vet $(VETPACKAGES)

.PHONY: fmt
fmt:
	@$(GOFMT) -w $(FILES)

.PHONY: fmt-check
fmt-check:
	@if [ -n "$$($(GOFMT) -d $(FILES))" ]; then \
		echo "Run 'make fmt' and commit the result:"; \
		exit 1; \
	fi;

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: test
test:
	go test ./... -count=1 -timeout 15s

test-race:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out
	go tool cover -func coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'
	rm coverage.out
