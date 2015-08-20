#!/usr/bin/env python
"""
Put the basic decks in the database

Deck sources: http://hearthstone.gamepedia.com/Basic_deck
"""
import sqlite3
import sys
from datetime import datetime


class DeckType:
	NORMAL_DECK = 1
	AI_DECK = 2
	DRAFT_DECK = 4
	PRECON_DECK = 5
	TAVERN_BRAWL_DECK = 6


BASIC_PRECON_DRUID = [
	"Boulderfist Ogre",
	"Chillwind Yeti",
	"Claw",
	"Core Hound",
	"Darkscale Healer",
	"Elven Archer",
	"Healing Touch",
	"Innervate",
	"Lord of the Arena",
	"Mark of the Wild",
	"Nightblade",
	"Oasis Snapjaw",
	"River Crocolisk",
	"Silverback Patriarch",
	"Wild Growth",
]

BASIC_PRECON_HUNTER = [
	"Arcane Shot",
	"Bloodfen Raptor",
	"Core Hound",
	"Houndmaster",
	"Ironforge Rifleman",
	"Multi-Shot",
	"Oasis Snapjaw",
	"Raid Leader",
	"Razorfen Hunter",
	"River Crocolisk",
	"Silverback Patriarch",
	"Stonetusk Boar",
	"Stormpike Commando",
	"Timber Wolf",
	"Tracking",
]

BASIC_PRECON_MAGE = [
	"Arcane Explosion",
	"Arcane Intellect",
	"Arcane Missiles",
	"Bloodfen Raptor",
	"Boulderfist Ogre",
	"Fireball",
	"Murloc Raider",
	"Nightblade",
	"Novice Engineer",
	"Oasis Snapjaw",
	"Polymorph",
	"Raid Leader",
	"River Crocolisk",
	"Sen'jin Shieldmasta",
	"Wolfrider",
]

BASIC_PRECON_PALADIN = [
	"Blessing of Might",
	"Elven Archer",
	"Gnomish Inventor",
	"Goldshire Footman",
	"Hammer of Wrath",
	"Hand of Protection",
	"Holy Light",
	"Ironforge Rifleman",
	"Light's Justice",
	"Lord of the Arena",
	"Novice Engineer",
	"Raid Leader",
	"Stormpike Commando",
	"Stormwind Champion",
	"Stormwind Knight",
]

BASIC_PRECON_PRIEST = [
	"Bloodfen Raptor",
	"Chillwind Yeti",
	"Core Hound",
	"Elven Archer",
	"Frostwolf Grunt",
	"Gurubashi Berserker",
	"Holy Smite",
	"Mind Blast",
	"Northshire Cleric",
	"Power Word: Shield",
	"Sen'jin Shieldmasta",
	"Shadow Word: Pain",
	"Shattered Sun Cleric",
	"Silverback Patriarch",
	"Voodoo Doctor",
]

BASIC_PRECON_ROGUE = [
	"Assassinate",
	"Backstab",
	"Bloodfen Raptor",
	"Deadly Poison",
	"Dragonling Mechanic",
	"Elven Archer",
	"Gnomish Inventor",
	"Goldshire Footman",
	"Ironforge Rifleman",
	"Nightblade",
	"Novice Engineer",
	"Sap",
	"Sinister Strike",
	"Stormpike Commando",
	"Stormwind Knight",
]

BASIC_PRECON_SHAMAN = [
	"Ancestral Healing",
	"Booty Bay Bodyguard",
	"Frost Shock",
	"Frostwolf Grunt",
	"Frostwolf Warlord",
	"Hex",
	"Murloc Raider",
	"Raid Leader",
	"Reckless Rocketeer",
	"Rockbiter Weapon",
	"Sen'jin Shieldmasta",
	"Stonetusk Boar",
	"Voodoo Doctor",
	"Windfury",
	"Wolfrider",
]

BASIC_PRECON_WARLOCK = [
	"Chillwind Yeti",
	"Darkscale Healer",
	"Drain Life",
	"Hellfire",
	"Kobold Geomancer",
	"Murloc Raider",
	"Ogre Magi",
	"Reckless Rocketeer",
	"River Crocolisk",
	"Shadow Bolt",
	"Succubus",
	"Voidwalker",
	"Voodoo Doctor",
	"War Golem",
	"Wolfrider",
]

BASIC_PRECON_WARRIOR = [
	"Boulderfist Ogre",
	"Charge",
	"Dragonling Mechanic",
	"Execute",
	"Fiery War Axe",
	"Frostwolf Grunt",
	"Gurubashi Berserker",
	"Heroic Strike",
	"Lord of the Arena",
	"Murloc Raider",
	"Murloc Tidehunter",
	"Razorfen Hunter",
	"Sen'jin Shieldmasta",
	"Warsong Commander",
	"Wolfrider",
]

BASIC_DECKS = {
	"Malfurion Stormrage": BASIC_PRECON_DRUID,
	"Rexxar": BASIC_PRECON_HUNTER,
	"Jaina Proudmoore": BASIC_PRECON_MAGE,
	"Uther Lightbringer": BASIC_PRECON_PALADIN,
	"Anduin Wrynn": BASIC_PRECON_PRIEST,
	"Valeera Sanguinar": BASIC_PRECON_ROGUE,
	"Thrall": BASIC_PRECON_SHAMAN,
	"Gul'dan": BASIC_PRECON_WARLOCK,
	"Garrosh Hellscream": BASIC_PRECON_WARRIOR,
}

def get_card_id(cursor, name):
	sql_select = "SELECT id FROM dbf_card WHERE name_enus = ? AND is_collectible = ?"
	rows = cursor.execute(sql_select, (name, True))
	rows = list(rows)
	assert len(rows) == 1
	return rows[0][0]


def main():
	if len(sys.argv) < 2:
		sys.stderr.write("USAGE: %s [dbfile]\n" % (sys.argv[0]))
		exit(1)
	dbfile = sys.argv[1]

	connection = sqlite3.connect(dbfile)
	cursor = connection.cursor()

	account_id = None
	deck_type = DeckType.PRECON_DECK
	last_modified = datetime.now()

	for hero_name, deck in BASIC_DECKS.items():
		hero_id = get_card_id(cursor, hero_name)
		hero_premium = False
		card_back_id = 0
		name = "Precon Basic %s" % (hero_name)
		print("Creating %s with cards %r" % (name, deck))
		cards = []
		cursor.execute("INSERT INTO deck VALUES (?, ?, ?, ?, ?, ?, ?, ?)", (
			None,
			account_id,
			deck_type,
			name,
			hero_id,
			hero_premium,
			card_back_id,
			last_modified,
		))
		deck_id = cursor.lastrowid
		assert deck_id

		for card in deck:
			id = None
			card_id = get_card_id(cursor, card)
			premium = False
			cards.append((id, deck_id, card_id, premium, 2))
		connection.executemany("INSERT INTO deck_card VALUES (?, ?, ?, ?, ?)", cards)

	# hardcode the arena cost for now. TODO: the rest of the store items
	product_type = 2
	pack_type = 0
	cost = 150
	cursor.execute("INSERT INTO product_gold_cost VALUES (?, ?, ?, ?)", (
		None,
		product_type,
		pack_type,
		cost
	))

	connection.commit()
	connection.close()


if __name__ == "__main__":
	main()
