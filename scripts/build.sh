#!/bin/bash
set -eufo pipefail

RED='\E[1;31m'
YELLOW='\E[1;33m'
RESET='\033[0m'

PROJECT_DIR="$(cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$PROJECT_DIR"

trap cleanup EXIT
cleanup() {
	if [ "$?" -ne 0 ]; then
		echo -e "${RED}BUILDING FAILED${RESET}"
	fi
}

go build -o bin/route53-ddns cmd/route53-ddns/main.go

echo -e "${YELLOW}Build completed: bin/route53-ddns${RESET}"
