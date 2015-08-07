#!/usr/bin/env python
"""
Create the default user and achievement tracking
"""
import sqlite3
import sys
from datetime import datetime


def main():
	if len(sys.argv) < 2:
		sys.stderr.write("USAGE: %s [dbfile]\n" % (sys.argv[0]))
		exit(1)
	dbfile = sys.argv[1]

	connection = sqlite3.connect(dbfile)
	cursor = connection.cursor()

	bnet_id = None
	updated_at = datetime.now()
	flags = None

	cursor.execute("INSERT INTO account VALUES (?, ?, ?, ?)", (
		None,
		bnet_id,
		updated_at,
		flags
	))

	account_id = cursor.lastrowid

	connection.commit()
	connection.close()


if __name__ == "__main__":
	main()
