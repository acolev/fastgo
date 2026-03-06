SWAG := $(shell sh -c 'GOBIN=$$(go env GOBIN); if [ -n "$$GOBIN" ]; then printf "%s/swag" "$$GOBIN"; else printf "%s/bin/swag" "$$(go env GOPATH)"; fi')
SWAG_VERSION := v1.16.6

.PHONY: docs swagger

docs:
	@test -x "$(SWAG)" || go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	@$(SWAG) init -g cmd/api/main.go -o docs --parseInternal
	@rm -f docs/docs.go

swagger: docs
