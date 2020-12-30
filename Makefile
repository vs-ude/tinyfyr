prefix ?= /usr
datadir ?= $(prefix)/share

compiler_binaries := tfc tfarch

export TFBASE = ${CURDIR}

SOURCE_DATE_EPOCH ?= $(shell date +%F\ %T)
buildFlags = -ldflags "-X 'main.buildDate=${SOURCE_DATE_EPOCH}'"

.PHONY: all
all: build

.PHONY: build
build: ${compiler_binaries}

.PHONY: ${compiler_binaries}
${compiler_binaries}:
	go build ${buildFlags} ./cmd/$@

.PHONY: run
run: tfc
	./tfc

.PHONY: test
test: test_go test_tf

.PHONY: test_go
test_go:
	@go test ./... || \
	(printf "\n\e[31mErrors occurred when running tests.\e[0m Please see above output for more information.\n\n" && exit 1) && \
	printf "\n\e[32mAll internal tests completed successfully.\e[0m\n\n"

.PHONY: test_tf
test_fyr: build clean_examples
	@./test/tf_code_tests.sh

.PHONY: clean
clean: clean_compiler clean_examples

.PHONY: clean_compiler
clean_compiler:
	rm -rf ${compiler_binaries}

.PHONY: clean_examples
clean_examples:
	find examples lib -type d -name 'pkg' -exec rm -rf {} +
	find examples -type d -name 'bin' -exec rm -rf {} +
