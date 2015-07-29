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
if [ ! -e hs-data ]; then
	git clone --depth=1 https://github.com/HearthSim/hs-data.git
fi
../scripts/dbf_to_sqlite.py ./hs-data ../pegasus.db
cd ..
rm -rf build
