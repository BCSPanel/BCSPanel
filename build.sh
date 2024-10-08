#!/bin/bash
set -e

cd frontend-antd
npm run build
cd ..

cd frontend-login2
npm run build
cd ..

echo '> go build'
go build
