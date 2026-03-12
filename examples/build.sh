#!/bin/bash

# Build and run the Morrohsu demo

echo "Building Morrohsu demo..."
cd ..

# Make sure we're in the right directory
if [ ! -d "pkg/tools" ]; then
    echo "Error: Cannot find pkg/tools directory"
    exit 1
fi

echo "Running Morrohsu demo..."
go run examples/morrohsu_demo.go

echo "Demo completed."