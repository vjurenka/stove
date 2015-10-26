#!/usr/bin/env python
import os
import sqlite3
import sys
from xml.etree import ElementTree


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
	primary_key = None
	for name, type in columns:
		cn = name.lower()
		if type == "LocString":
			cn += "_enus"
		t = "%s %s" % (cn, types[type])
		if cn in ("id", "unique_id") and not primary_key:
			t += " primary key"
			primary_key = cn
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


def get_tag(e, tag):
	elem = e.find('Tag[@enumID="%i"]' % (tag))
	if elem is not None:
		return int(elem.attrib["value"])
	return 0


def get_crafting_values(card_set, rarity):
	if card_set not in (3, 13, 15):
		return 0, 0, 0, 0
	elif rarity == 1:
		return 40, 5, 400, 50
	elif rarity == 3:
		return 100, 20, 800, 100
	elif rarity == 4:
		return 400, 100, 1600, 400
	elif rarity == 5:
		return 1600, 400, 3200, 1600
	return 0, 0, 0, 0

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
		tablename = os.path.splitext(path)[0].lower()
		with open(os.path.join(datadir, "DBF", path), "r") as f:
			xml = ElementTree.parse(f)

			cols = [(e.attrib["name"], e.attrib["type"]) for e in xml.findall("Column")]

			_columns = generate_columns(tablename, cols)
			create_tbl = "CREATE TABLE IF NOT EXISTS dbf_%s (%s)" % (tablename, _columns)
			connection.execute(create_tbl)

			values = []
			for record in xml.findall("Record"):
				fields = [get_field(tablename, record, column, type) for column, type in cols]

				values.append(fields)

			values_ph = ", ".join("?" for c in cols)
			insert_into = "INSERT INTO dbf_%s VALUES (%s)" % (tablename, values_ph)

			print(insert_into)

			connection.executemany(insert_into, values)

	# Add card names
	connection.execute("ALTER TABLE dbf_card ADD COLUMN name_enus text")

	# Add card class
	connection.execute("ALTER TABLE dbf_card ADD COLUMN class_id int")

	# Add Rarity
	connection.execute("ALTER TABLE dbf_card ADD COLUMN rarity int")

	# Add Card Set
	connection.execute("ALTER TABLE dbf_card ADD COLUMN card_set int")

	# Add crafting values
	connection.execute("ALTER TABLE dbf_card ADD COLUMN buy_price int")
	connection.execute("ALTER TABLE dbf_card ADD COLUMN sell_price int")
	connection.execute("ALTER TABLE dbf_card ADD COLUMN gold_buy_price int")
	connection.execute("ALTER TABLE dbf_card ADD COLUMN gold_sell_price int")

	cur = connection.cursor()
	cur.execute("SELECT id, note_mini_guid FROM dbf_card")
	rows = cur.fetchall()

	with open(os.path.join(datadir, "CardDefs.xml"), "r") as f:
		xml = ElementTree.parse(f)

		for pk, id in rows:
			xpath = 'Entity[@CardID="%s"]' % (id)
			e = xml.find(xpath)
			if e is None:
				print("WARNING: Could not find card %r in hs-data." % (id))
				continue
			name = e.find('Tag[@enumID="185"]/enUS').text
			connection.execute("UPDATE dbf_card SET name_enus = ? WHERE id = ?", (name, pk))
			card_set = get_tag(e, 183)
			card_class = get_tag(e, 199)
			rarity = get_tag(e, 203)
			craft_cost, de_cost, gold_craft_cost, gold_de_cost = get_crafting_values(card_set, rarity)
			vars = card_class, rarity, card_set, craft_cost, de_cost, gold_craft_cost, gold_de_cost, pk
			connection.execute("""UPDATE dbf_card SET
				class_id = ?,
				rarity = ?,
				card_set = ?,
				buy_price = ?,
				sell_price = ?,
				gold_buy_price = ?,
				gold_sell_price = ?
			WHERE id = ?""", vars)

	connection.commit()
	connection.close()


if __name__ == "__main__":
	main()
