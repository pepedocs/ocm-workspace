.PHONY: build
build:
	go build -o workspace

.PHONY: install
install:
	go install ocm-workspace

.PHONY: buildImage
buildImage:
	./workspace build

.PHONY: all
all: build install buildImage