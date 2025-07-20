.PHONY: help
help: ## Display available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z0-9_-]+:.*## .*$$' Makefile | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the docker image
	@echo "Running 'make $@'..."
	docker build . -t gateway-purch:latest 

down: ## Bring down the docker container
	@echo "Running 'make $@'..."
	docker stop gateway-purch
	docker rm gateway-purch

up: ## Run the docker image
	@echo "Running 'make $@'..."
	docker run -p 8080:8080 --name gateway-purch -d gateway-purch:latest 