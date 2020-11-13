SHELL             := /usr/bin/env bash
PROJECT_NAME      ?= sample-app
ENV               ?=
export ENV

##############################
# Help
##############################

.DEFAULT_GOAL:=help

.PHONY: help
help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# aux: some things for printing color messages
bold     := $(shell tput bold)
sgr0     := $(shell tput sgr0)
ccred    := $(shell tput setaf 001)
ccgreen  := $(shell tput setaf 002)
ccyellow := $(shell tput setaf 003)

.PHONY: build
all: build

##############################################
##@ Local development
##############################################

format: fmt
fmt: ## format the Go source files
	for d in . ; do printf ">>> $(bold)$(ccgreen)Formatting files in $$d$(sgr0)\n" ;  go fmt $$d || /bin/true ; done

run: ## run the sample-app
	go run main.go

##############################################
##@ Docker images
##############################################

build: ## build the k8s-initializer-sample-app:latest image
	@printf '>>> $(bold)$(ccgreen)Building sample-app image...$(sgr0)\n'
	docker build -t k8s-initializer-sample-app -f Dockerfile .

push-image: ## push the k8s-initializer-sample-app:latest image to grc.io/datawire
	@printf '>>> $(bold)$(ccgreen)Tagging sample-app image...$(sgr0)\n'
	docker tag k8s-initializer-sample-app:latest gcr.io/datawire/k8s-initializer-sample-app:latest
	@printf '>>> $(bold)$(ccgreen)Pushing sample-app image...$(sgr0)\n'
	docker push gcr.io/datawire/k8s-initializer-sample-app:latest