#!/bin/bash
set -e

BASEDIR="$(realpath $(dirname $0))"
DATADIR="$BASEDIR/hs-data"
DBFILE="$BASEDIR/db/pegasus.db"
HSDATA="https://github.com/HearthSim/hs-data.git"

mkdir -p "$BASEDIR/db"

if [ -e "$DBFILE" ]; then
	echo "$DBFILE already exists.  If you want to rerun the bootstrapper, remove it."
	exit 1
fi

echo "Fetching data files from $HSDATA"
if [ ! -e "$DATADIR" ]; then
	git clone --depth=1 "$HSDATA" "$DATADIR"
else
	git -C "$DATADIR" pull >/dev/null
fi

echo "Creating database in $DBFILE"
"$BASEDIR/scripts/dbf_to_sqlite.py" "$DATADIR" "$DBFILE"

echo "Creating pegasus tables"
go run "$BASEDIR/migrate.go"

echo "Done."
