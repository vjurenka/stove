#!/bin/bash
set -e

cd "$(dirname $0)"

if [ -e pegasus.db ]; then
	echo "pegasus.db already exists.  If you want to rerun the bootstrapper," \
	"then remove pegasus.db."
	exit 1
fi

mkdir -p build
cd build
if [ ! -e fireplace ]; then
	git clone --depth 1 --recursive https://github.com/jleclanche/fireplace.git
fi
cd fireplace
source ./bootstrap.sh
cp ../../scripts/dbf_to_sqlite.py .
./dbf_to_sqlite.py ./data ../../pegasus.db
cd ../../
rm -r build
