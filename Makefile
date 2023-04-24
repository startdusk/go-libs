.PHONY: test
test: clean
	@go test ./...

.PHONY: clean
clean:
	@go mod tidy
	@go vet ./...
	@go fmt ./...

