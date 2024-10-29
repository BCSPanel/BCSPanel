#!/bin/bash
set -e

cd frontend-antd
npm run build
cd ..

cd frontend-login
npm run build
cd ..

echo '> buildgo'
./buildgo.sh
