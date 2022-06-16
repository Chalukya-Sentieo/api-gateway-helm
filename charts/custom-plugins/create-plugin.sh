#!/bin/sh
mkdir "$1"
cd "$1"
# UNCOMMENT `go mod {}` commands to enable go.mod and go.sum file generation
#go mod init "$1"
touch "$1.go"
#echo -e "package main\n\nimport (\n\t\"github.com/Kong/go-pdk\"\n)\n\ntype Config struct {\n}" >> "$1.go"
tee -a "$1.go" <<EOF
package main

import (
	"github.com/Kong/go-pdk"
)

type Config struct {
}
EOF
# UNCOMMENT `go mod {}` commands to enable go.mod and go.sum file generation
#go mod tidy
cd ..