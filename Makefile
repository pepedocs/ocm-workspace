.PHONY: build
build:
	go build -o workspace

.PHONY: install
install:
	go install ocm-workspace

.PHONY: buildImage
buildImage:
ifeq ($(GITHUB_ACTIONS),)
	./workspace build
else
	$(shell echo "myVersion: user" >> ~/.ocm-workspace.yaml)
	./workspace build
endif

.PHONY: all
all: build install buildImage