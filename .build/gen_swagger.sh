#!/bin/bash

# Run from project root

set -e

swag init --dir . --output ./build --generalInfo ./cmd/dnode/main.go

swagger-go serve -F=swagger ./build/swagger.yaml