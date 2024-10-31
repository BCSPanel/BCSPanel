#!/bin/bash
set -e

git submodule update --init --remote

cd frontend
git checkout main
npm i
cd ..

go get
