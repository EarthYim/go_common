SHELL := /bin/bash

.PHONY: dev
dev:
	$(call print-target)
	docker-compose build
	docker-compose up

.PHONY: cleanup
cleanup:
	docker-compose down