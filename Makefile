ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

cli: check_lfs
	go build -o bin/livekit-cli ./cmd/livekit-cli
	GOOS=linux GOARCH=amd64 go build -o bin/livekit-cli-linux ./cmd/livekit-cli

install: cli
	cp bin/livekit-cli $(GOBIN)/

check_lfs:
	@{ \
	if [ ! -n $(find pkg/provider/resources -name neon_720_2000.ivf -size +100) ]; then \
		echo "Video resources not found. Ensure Git LFS is installed"; \
		exit 1; \
	fi \
	}
