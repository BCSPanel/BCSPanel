#!/bin/bash
set -e

git submodule update --init --remote

cd frontend-antd
git checkout main
npm i
cd ..

cd frontend-login2
git checkout master
npm i
cd ..
