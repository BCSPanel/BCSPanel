#!/bin/bash
set -e

git submodule update --init --remote

cd frontend
git checkout master
npm i
cd ..
