#!/usr/bin/env bash
set -euo pipefail

# ---------------------------
# ANSI colors (printf-based)
# ---------------------------
if [[ -t 1 ]]; then
  ESC=$'\033'
  BOLD="${ESC}[1m"; RESET="${ESC}[0m"
  YELLOW_BG="${ESC}[43m"; GREEN_BG="${ESC}[42m"; RED_BG="${ESC}[41m"
else
  BOLD=""; RESET=""; YELLOW_BG=""; GREEN_BG=""; RED_BG=""
fi

banner() { # banner "<bg_color>" "Message..."
  local bg="$1"; shift
  printf "\n%b%b%s%b\n" "$bg" "$BOLD" "$*" "$RESET"
}

say() { printf "%s\n" "$*"; }
hr()  { printf "%s\n" "------------------------------------------------------------"; }
ts()  { date +"%Y-%m-%d %H:%M:%S %Z"; }

# ---------------------------
# Config
# ---------------------------
BIN_NAME="klyntar"
TARGET_OS="${1:-$(go env GOOS)}"
TARGET_ARCH="${2:-$(go env GOARCH)}"
SUCCESS_ART="./images/success_build.txt"
FAIL_ART="./images/fail_build.txt"

# ---------------------------
# Error handler
# ---------------------------
on_error() {
  banner "$RED_BG" "Build failed  •  $(ts)"
  if [[ -f "$FAIL_ART" ]]; then
    printf "\n"
    cat "$FAIL_ART"
    printf "\n"
  fi
}
trap on_error ERR

# ---------------------------
# Build
# ---------------------------
START_TS=$(date +%s)

banner "$YELLOW_BG" "Fetching dependencies  •  $(ts)"
hr
say "Working dir : $(pwd)"
say "Go version  : $(go version | awk '{print $3, $4}')"
say "Target      : ${TARGET_OS}/${TARGET_ARCH}"
hr
go mod download

banner "$GREEN_BG" "Core building process started  •  $(ts)"
say "${BOLD}Building the project for ${TARGET_OS}/${TARGET_ARCH}...${RESET}"
hr
GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -o "$BIN_NAME" .

# ---------------------------
# Success
# ---------------------------
END_TS=$(date +%s)
ELAPSED=$((END_TS - START_TS))

banner "$GREEN_BG" "Build succeeded  •  $(ts)"
say "Binary       : ${BIN_NAME}"
say "Target       : ${TARGET_OS}/${TARGET_ARCH}"
say "Output path  : $(pwd)/${BIN_NAME}"
say "Elapsed time : ${ELAPSED}s"
hr

if [[ -f "$SUCCESS_ART" ]]; then
  printf "\n"
  cat "$SUCCESS_ART"
  printf "\n"
fi
