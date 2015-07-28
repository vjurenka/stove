#!/usr/bin/env python
import os
import sqlite3
import sys
from xml.etree import ElementTree
from fireplace import cards


def generate_columns(table, columns):
	ret = []
	types = {
		"Bool": "bool",
		"Int": "int",
		"Long": "long",
		"ULong": "unsigned long",
		"String": "text",
		"LocString": "text",
		"AssetPath": "text",
	}
	for name, type in columns:
		cn = name.lower()
		if type == "LocString":
			cn += "_enus"
		t = "%s %s" % (cn, types[type])
		if cn == "id":
			t += " primary key"
		ret.append(t)

	return ",\n".join(ret)


def get_field(table, record, column, type):
	xpath = './Field[@column="%s"]' % (column)
	if type == "LocString":
		xpath += "/enUS"

	data = record.find(xpath)
	if data is None:
		return None
	data = data.text

	if type == "Bool":
		return True if data == "True" else False
	elif type in ("Int", "Long", "ULong"):
		if data is None:
			return None
		return int(data)
	return data


def main():
	if len(sys.argv) < 3:
		sys.stderr.write("USAGE: %s [datadir] [dbfile]\n" % (sys.argv[0]))
		exit(1)
	datadir = sys.argv[1]
	dbfile = sys.argv[2]

	connection = sqlite3.connect(dbfile)

	files = [
		"ACHIEVE.xml",
		"ADVENTURE.xml",
		"ADVENTURE_DATA.xml",
		"ADVENTURE_MISSION.xml",
		"BANNER.xml",
		"BOARD.xml",
		"BOOSTER.xml",
		"CARD_BACK.xml",
		"CARD.xml",
		"FIXED_REWARD.xml",
		"FIXED_REWARD_ACTION.xml",
		"FIXED_REWARD_MAP.xml",
		"HERO.xml",
		"SCENARIO.xml",
		"SEASON.xml",
		"WING.xml",
	]

	for path in files:
		tablename = os.path.splitext(path)[0]
		with open(os.path.join(datadir, "DBF", path), "r", encoding="utf8") as f:
			xml = ElementTree.parse(f)

			cols = [(e.attrib["name"], e.attrib["type"]) for e in xml.findall("Column")]

			_columns = generate_columns(tablename, cols)
			create_tbl = "CREATE TABLE IF NOT EXISTS %s (%s)" % (tablename, _columns)
			connection.execute(create_tbl)

			values = []
			for record in xml.findall("Record"):
				fields = [get_field(tablename, record, column, type) for column, type in cols]
				if tablename == "card":
					fields.append(getattr(cards, fields[1]).name)

				values.append(fields)

			values_ph = ", ".join("?" for c in cols)
			insert_into = "INSERT INTO %s VALUES (%s)" % (tablename, values_ph)

			print(insert_into)

			connection.executemany(insert_into, values)

	# Add card names
	connection.execute("ALTER TABLE card ADD COLUMN name_enus text")
	cur = connection.cursor()
	cur.execute("SELECT id, note_mini_guid FROM card")
	rows = cur.fetchall()

	for pk, id in rows:
		name = getattr(cards, id).name
		connection.execute("UPDATE card SET name_enus = ? WHERE id = ?", (name, pk))

	connection.commit()
	connection.close()



if __name__ == "__main__":
	main()
