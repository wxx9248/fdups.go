# Go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_VET=$(GO_CMD) vet
GO_TEST=$(GO_CMD) test

# Binary names
BINARY_NAME=fdups

# Directories
DIST_DIR=dist

# Targets
debug:
	$(GO_VET) -tags="debug"
	$(GO_TEST) -tags="debug"
	$(GO_BUILD) -tags="debug" -o $(DIST_DIR)/$(BINARY_NAME)-debug .

release:
	$(GO_VET) -tags="release"
	$(GO_TEST) -tags="release"
	$(GO_BUILD) -tags="release" -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-release .

clean:
	$(GO_CLEAN)
	rm -rf $(DIST_DIR)
