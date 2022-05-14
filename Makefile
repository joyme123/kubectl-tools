OUTPUT_DIR=./bin
NAME=kubectl-tools
VERSION_TMP  ?= $(shell git describe --tags --always --dirty)
VERSION      = $(shell echo ${VERSION_TMP} | sed -e "s/\//-/g" )

.PHONY: build

iputils:
	docker run -it --rm -v ${PWD}/bin/tools:/app ghcr.io/joyme123/gcc:4.9 sh -c \
	'git clone https://github.com/dgibson/iputils \
	&& cd iputils && make LDFLAGS="-static -s" USE_GNUTLS="no" \
	&& mv ping /app && mv ping6 /app && mv arping /app \
	&& mv tracepath /app && mv tracepath6 /app && mv traceroute6 /app'


build: 
	go build -o ${OUTPUT_DIR}/${NAME} 								   \
	-ldflags "-s -w -X $(ROOT)/pkg/version.module=$(NAME)              \
	-X $(ROOT)/pkg/version.branch=$(BRANCH)                            \
	-X $(ROOT)/pkg/version.gitCommit=$(GITCOMMIT)                      \
	-X $(ROOT)/pkg/version.gitTreeState=$(GITTREESTATE)                \
	-X $(ROOT)/pkg/version.buildDate=$(BUILDDATE)                     \
	-X $(ROOT)/pkg/version.version=$(VERSION)" 						   \
	.

package-tools:
	cd bin/tools && tar -czf ../../tools.tar.gz *