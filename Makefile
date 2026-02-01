SHELL := /bin/bash

.PHONY: dev cleanup key build build-prod
dev:
	$(call print-target)
	docker-compose build
	docker-compose up

cleanup:
	docker-compose down

key:
	openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem
	openssl ec -in private_key.pem -pubout -out public_key.pem
	base64 -i private_key.pem -o private_key_64.txt
	base64 -i public_key.pem -o public_key_64.txt

	.PHONY: hash-password help

# Default target
help:
	@echo "Usage:"
	@echo "  make hash-password PASS=yourpassword"
	@echo ""
	@echo "Example:"
	@echo "  make hash-password PASS=mysecret123"

# Hash password using SHA256
hash-password:
	@if [ -z "$(PASS)" ]; then \
		echo "Error: PASS variable is required"; \
		echo "Usage: make hash-password PASS=yourpassword"; \
		exit 1; \
	fi
	@echo "Hashing password: $(PASS)"
	@echo -n "$(PASS)" | sha256sum | awk '{print $$1}'
	@echo ""
	@echo "Add this to your config/env:"
	@echo "ADMIN_PASSWORD=$$(echo -n '$(PASS)' | sha256sum | awk '{print $$1}')"

build-prod:
	docker buildx build \
	--platform=linux/amd64 \
	-t asia-southeast1-docker.pkg.dev/popclick-485102/popclick-server/popclick-server:latest \
	.
