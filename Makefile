bin:
	CGO_ENABLED=0 go build -o tavern ./cmd/tavern

.PHONY: bin
