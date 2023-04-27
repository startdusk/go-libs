.PHONY: test
test: clean
	@go test ./... -v

.PHONY: clean
clean:
	@go mod tidy
	@go vet ./...
	@go fmt ./...

