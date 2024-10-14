#!/bin/bash
set -e

go build -trimpath -ldflags "-w -s"
