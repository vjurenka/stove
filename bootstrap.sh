#!/bin/bash
set -e

if [ -z "$PEGASUS_DB" ]; then
	echo >&2 "Please use the PEGASUS_DB variable to set the location of the db file"
	exit 1
fi

BASEDIR="$(readlink -f $(dirname $0))"
DATADIR="$BASEDIR/hs-data"
HSDATA="https://github.com/HearthSim/hs-data.git"

mkdir -p "$BASEDIR/db"

if [ -e "$PEGASUS_DB" ]; then
	echo >&2 "$PEGASUS_DB already exists.  If you want to rerun the bootstrapper, remove it."
	exit 1
fi

if [ -z "$BNET_DB" ]; then
	echo >&2 "BNET_DB is not defined, using default location $BASEDIR/db/bnet.db"
	BNET_DB="$BASEDIR/db/bnet.db"
fi

if [ -e "$BNET_DB" ]; then
	echo >&2 "$BNET_DB already exists.  If you want to rerun the bootstrapper, remove it."
	exit 1
fi

echo "Fetching data files from $HSDATA"
if [ ! -e "$DATADIR" ]; then
	git clone --depth=1 "$HSDATA" "$DATADIR"
else
	git -C "$DATADIR" pull >/dev/null
fi

echo "Creating database in $PEGASUS_DB"
"$BASEDIR/scripts/dbf_to_sqlite.py" "$DATADIR" "$PEGASUS_DB"

echo "Creating pegasus tables"
go run "$BASEDIR/stove.go" -migrate

echo "Initializing database"
"$BASEDIR/scripts/initialize_database.py" "$PEGASUS_DB"

echo "Creating default user"
"$BASEDIR/scripts/create_default_user.py" "$PEGASUS_DB" "$BNET_DB"

echo "Done."
