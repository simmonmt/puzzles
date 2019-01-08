package game

import (
	"character"
	"logger"
	"spells"
)

func bossAttack(player *character.Player, boss *character.Boss) {
	damage := boss.Damage - player.Armor
	if damage < 1 {
		damage = 1
	}
	logger.LogF("Boss attacks for %v (boss damage %v, player armor %v)\n", damage, boss.Damage, player.Armor)
	player.HP -= damage
}

func nextSpell(spellProvider spells.Provider, availMana int) spells.Spell {
	spell := spellProvider.NextSpell(availMana)
	if spell == nil {
		return nil
	}
	if spell.Cost() > availMana {
		panic("expensive spell")
	}
	if spell.IsActive() {
		panic("already active")
	}
	return spell
}

func Run(player character.Player, boss character.Boss, allSpells []spells.Spell, spellProvider spells.Provider) (bossDead bool, manaUsed int, spellsCast []spells.Spell) {
	manaUsed = 0
	spellsCast = []spells.Spell{}

	for i := 0; ; i++ {
		if i > 0 {
			logger.LogLn()
		}

		playerTurn := i%2 == 0

		if playerTurn {
			logger.LogLn("-- Player turn")
		} else {
			logger.LogLn("-- Boss turn")
		}

		player.Print()
		boss.Print()

		if playerTurn {
			player.HP--
			logger.LogF("Player initial HP drain, now %v\n", player.HP)

			if player.HP <= 0 {
				logger.LogLn("Initial HP drain caused player death.")
			} else {
				for _, spell := range allSpells {
					if spell.IsActive() {
						spell.TurnStart(&player, &boss)
					}
				}

				spellToCast := nextSpell(spellProvider, player.Mana)
				if spellToCast == nil {
					logger.LogF("Player mana %v too low; killing player.", player.Mana)
					player.HP = 0
				} else {
					player.Mana -= spellToCast.Cost()
					manaUsed += spellToCast.Cost()
					spellsCast = append(spellsCast, spellToCast)
					spellToCast.Activate(&player, &boss)
				}
			}
		} else { // boss turn
			for _, spell := range allSpells {
				if spell.IsActive() {
					spell.TurnStart(&player, &boss)
				}
			}

			if boss.HP > 0 {
				bossAttack(&player, &boss)
			}
		}

		if player.HP <= 0 {
			logger.LogLn("Player is dead.")
			bossDead = false
			return
		} else if boss.HP <= 0 {
			logger.LogLn("Boss is dead.")
			bossDead = true
			return
		}
	}

	panic("unreached")
}
