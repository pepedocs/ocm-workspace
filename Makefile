.PHONY: build
build:
	go build -o workspace

.PHONY: install
install:
	go install ocm-workspace

.PHONY: buildImage
buildImage:
	echo "this config is for the build" >> ~/.ocm-workspace.yaml &&
	./workspace build

.PHONY: all
all: build install buildImage