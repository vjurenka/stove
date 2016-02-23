#!/usr/bin/env python
"""
Create the default user and achievement tracking
"""
import sqlite3
import sys
from datetime import datetime


def main():
	if len(sys.argv) < 3:
		sys.stderr.write("USAGE: %s [dbfile] [bnet dbfile]\n" % (sys.argv[0]))
		exit(1)
	dbfile = sys.argv[1]
	dbfile_bnet = sys.argv[2]

	connection = sqlite3.connect(dbfile)
	cursor = connection.cursor()

	bnet_id = None
	now = datetime.now()
	updated_at = now
	flags = None
	dust = 100000
	gold = 2000

	cursor.execute("INSERT INTO account VALUES (?, ?, ?, ?, ?, ?)", (
		None,
		bnet_id,
		gold,
		dust,
		updated_at,
		flags
	))
	account_id = cursor.lastrowid
	assert account_id

	achieves = []
	for dbf_achieve in cursor.execute("SELECT * FROM dbf_achieve"):
		id = None
		achieve_id = dbf_achieve[0]
		progress = 1
		ack_progress = 1
		completion_count = 1
		active = False
		date_given = now
		date_completed = now

		achieves.append((
			id,
			account_id,
			achieve_id,
			progress,
			ack_progress,
			completion_count,
			active,
			date_given,
			date_completed
		))

	connection.executemany("INSERT INTO achieve VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", achieves)

	values = []
	for card_id, class_id in cursor.execute("SELECT dbf_card.id, dbf_card.class_id FROM dbf_card " + \
		"JOIN dbf_hero ON dbf_card.note_mini_guid = dbf_hero.card_id WHERE not dbf_hero.store_bought"):
		id = None
		premium = False
		values.append((id, account_id, class_id, card_id, premium))
	connection.executemany("INSERT INTO favorite_hero VALUES (?, ?, ?, ?, ?)", values)

	connection.commit()
	connection.close()

	# create bnet.db with default account (based on stove/wiki)
	connection_bnet = sqlite3.connect(dbfile_bnet)
	cursor_bnet = connection_bnet.cursor()
	cursor_bnet.execute("INSERT INTO account VALUES (?, ?, ?, ?, ?)", (
		None,
		"test@hearthsim.info",
		"0123456789abcdef0123456789abcdef",
		"Test#1234",
		flags
	))
	account_bnet_id = cursor_bnet.lastrowid
	assert account_bnet_id

	connection_bnet.commit()
	connection_bnet.close()

if __name__ == "__main__":
	main()
