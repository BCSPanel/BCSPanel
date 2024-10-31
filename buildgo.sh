#!/bin/bash
set -e

go build -trimpath -ldflags "-w -s"

cd command
go build -trimpath -ldflags "-w -s"
cd ..
