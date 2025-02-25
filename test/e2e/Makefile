# [AGORIC] Need this to download the "latest" Docker image of this minor version.
DOCKER_LATEST_TAG := v0.37.x

all: docker generator runner

docker:
	@echo "Building E2E Docker image"
	@docker build \
		--tag cometbft/e2e-node:local-version \
		-f docker/Dockerfile ../..
	docker pull cometbft/e2e-node:$(DOCKER_LATEST_TAG)
	docker tag cometbft/e2e-node:$(DOCKER_LATEST_TAG) cometbft/e2e-node:latest

# We need to build support for database backends into the app in
# order to build a binary with a CometBFT node in it (for built-in
# ABCI testing).
node:
	go build -o build/node -tags badgerdb,boltdb,cleveldb,rocksdb ./node

generator:
	go build -o build/generator ./generator

runner:
	go build -o build/runner ./runner

.PHONY: all node docker generator runner 
