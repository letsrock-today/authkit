.PHONY: \
	all \
	clean \
	test \
	build \
	generate

all: \
	clean \
	test \
	build
	@$(MAKE) clean --no-print-directory
	@echo "All Done."

clean:
	@echo "Clean"
	@glide nv | xargs go clean -i -r
	@find ./authkit/mocks/ -type f -path *.go ! -name doc.go -delete
	@find ./sample/ -type f -path *_generated.go -delete
	@echo "Clean Done."

generate:
	@echo "Generate"
	@glide nv | xargs go generate
	@echo "Generate Done."

test: \
	generate
	@echo "Test"
	@glide nv | xargs go test
	@-glide nv | xargs go vet
	@-glide nv | xargs -L 1 golint
	@echo "Test Done."

build: \
	generate
	@echo "Build"
	@glide nv | xargs go build
	@echo "Build Done."
	@echo "Note: this build only checks if it can be done, it does not preserve output files."
	@echo "Note: use make in samples' dirs to build executables."
