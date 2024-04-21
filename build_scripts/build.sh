#!/bin/bash

BIN_NAME="klyntar"

# Build the core

echo -e "\e[43mBuilding Klyntar...\e[49m"

go build -o "$BIN_NAME" ./main.go

# Check if OK

if [ $? -eq 0 ]; then

    echo "KLY was successfully built"
    
    
    CURRENT_DIR="$(pwd)"
    
    # Full path to binary
    BIN_PATH="$CURRENT_DIR/$BIN_NAME"
    
    echo "Adding KLY to PATH..."
    echo "export PATH=\"\$PATH:$BIN_PATH\"" >> ~/.bashrc
    source ~/.bashrc
    
    cat ../images/success_build.txt
    
else
    cat ../images/fail_build.txt
fi
