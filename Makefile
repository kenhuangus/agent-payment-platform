SHELL := /bin/bash

.PHONY: all bootstrap build test openapi avro

all: build

bootstrap:
	@echo "Bootstrap placeholder (lint/tools install)"

build:
	@echo "Building services..."
	@cd services/identity && go build ./... || true

openapi:
	@echo "OpenAPI validation placeholder" && if exist libs\openapi\openapi.yaml echo "OpenAPI file found"

avro:
	@echo "Avro schemas present" && dir libs\common-proto-avro-schemas\*.avsc
