#!/bin/bash

BIN_NAME="klyntar"


###################################################
#           Install all dependencies              #
###################################################

echo -e "\e[43mFetching dependencies ...\e[49m"

go get ./...

echo -e "\e[42mCore building process started\e[49m"

###################################################
#               Building the core                 #
###################################################

# Build the core

go build -o "$BIN_NAME" ./main.go

# Check if OK

if [ $? -eq 0 ]; then

    echo "Core was successfully built"
    
    
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
