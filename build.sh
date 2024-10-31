#!/bin/bash
set -e

cd frontend
npm run build
cd ..

echo '> buildgo'
./buildgo.sh
