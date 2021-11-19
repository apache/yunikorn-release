.PHONY: build clean

TARGET_DIR ?= target
BINARY_DIR ?= perf-tools-bin
BINARY ?= perf-tools

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build: clean
	@mkdir ${TARGET_DIR}
	@mkdir ${TARGET_DIR}/${BINARY_DIR}
	@echo "[Action] mkdir target and bin directory"
	@CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${TARGET_DIR}/${BINARY_DIR}/${BINARY} .
	@echo "[Action] build binary"
	@cp conf.yaml ${TARGET_DIR}/${BINARY_DIR}
	@echo "[Action] copy binary and conf file to binary directory"
	@tar -zcvf ${TARGET_DIR}/${BINARY_DIR}.tar.gz -C ${TARGET_DIR} ${BINARY_DIR}
	@echo "[Action] generate binary package: ${TARGET_DIR}/${BINARY_DIR}.tar.gz"

clean:
	@if [ -d "${TARGET_DIR}" ] ; then rm -rf ${TARGET_DIR} && echo "[Action] cleaned target directory" ; fi
