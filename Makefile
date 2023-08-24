.PHONY: build
build:
	go build -o workspace

.PHONY: install
install:
	go install ocm-workspace

.PHONY: buildImage
buildImage:
	$(shell echo "myVersion: user" >> ~/.ocm-workspace.yaml) \
	./workspace build

.PHONY: all
all: build install buildImage