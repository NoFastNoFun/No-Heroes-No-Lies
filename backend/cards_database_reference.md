# No Heroes No Lies - Cards Database Reference

## Hero Cards (15 Total)

| ID | Name | Strength | Default Spawn | Rarity | Power 1 | Power 2 | Notes |
|---|---|---|---|---|---|---|---|
| `troll` | Troll | 7 | 2 | Uncommon | fight_monster | change_monster | Combat specialist |
| `esprit` | Esprit Sylvestre | 6 | 2 | Uncommon | fight_player_with_discarded_card | add_strength_to_challenger | PvP fighter |
| `barde` | Barde | 2 | 1 | Rare | shuffle_heroes_deck | change_hero | Deck manipulation |
| `elf` | Elf | 8 | 2 | Uncommon | exchange_gems_with_player | exchange_2coins_for_life | Resource trading |
| `nain` | Nain | 5 | 2 | Uncommon | gain_2gems | gain_life | Resource generation |
| `mage` | Mage | 9 | 2 | Uncommon | see_player_card | remove_coin | Information & disruption |
| `exorciste` | Exorciste | 3 | 4 | Common | shoot_player | all_in | High risk/reward |
| `gobelin` | Gobelin | 4 | 2 | Uncommon | steal_gem | steal_gem | Gem thief (dual steal) |
| `geant` | Géant | 10 | 1 | Rare | keep_gems (passive) | parry_this_you_f_casual | Tank with damage |
| `chevalier` | Chevalier | 11 | 1 | Rare | teamwork (passive) | steal_coin | Elite knight |
| `shapeshifter` | Shapeshifter | 1 | 1 | Rare | mimic_hero | mimic_power | Copycat abilities |
| `werewolf` | Werewolf | 0* | 2 | Uncommon | alternate_strength (passive) | dual_attack | *Alternates 5/12 |
| `witch` | Witch | 8 | 1 | Rare | blind_draw | force_transform | Transformation magic |
| `prince` | Mad Prince | 13 | 1 | Rare | tribute (passive) | execution | Strongest hero |
| `queen` | Mad Queen | 9 | 1 | Rare | royal_immunity (passive) | command_the_dead | Immune to attacks |

### Hero Strength Tiers
- **Legendary (11-13):** Mad Prince (13), Chevalier (11)
- **Elite (9-10):** Géant (10), Mage (9), Mad Queen (9)
- **Strong (7-8):** Troll (7), Elf (8), Witch (8)
- **Moderate (3-6):** Nain (5), Esprit (6), Gobelin (4), Exorciste (3)
- **Special (0-2):** Werewolf (5/12 alternating), Shapeshifter (1), Barde (2)

---

## Monster Cards (18 Total)

### Common Monsters (4 cards - 49.5% spawn rate)
| ID | Name | Strength | Loot Coins | Loot Gems | Rarity |
|---|---|---|---|---|---|
| `skeleton` | Skeleton | 5 | 1 | 1 | common |
| `goblin_warrior` | Goblin Warrior | 5 | 1 | 1 | common |
| `wyvern` | Wyvern | 6 | 1 | 2 | common |
| `minotaur` | Minotaur | 6 | 1 | 2 | common |

### Uncommon Monsters (5 cards - 30% spawn rate)
| ID | Name | Strength | Loot Coins | Loot Gems | Rarity |
|---|---|---|---|---|---|
| `spectre` | Spectre | 7 | 1 | 3 | uncommon |
| `basilisk` | Basilisk | 7 | 1 | 3 | uncommon |
| `golem` | Stone Golem | 8 | 1 | 3 | uncommon |
| `griffon` | Griffon | 8 | 1 | 4 | uncommon |
| `banshee` | Banshee | 7 | 1 | 4 | uncommon |

### Rare Monsters (4 cards - 15% spawn rate)
| ID | Name | Strength | Loot Coins | Loot Gems | Rarity |
|---|---|---|---|---|---|
| `dragon` | Ancient Dragon | 9 | 1 | 5 | rare |
| `chimera` | Chimera | 9 | 1 | 5 | rare |
| `lich` | Lich King | 10 | 1 | 6 | rare |
| `frost_giant` | Frost Giant | 10 | 1 | 5 | rare |

### Legendary Monsters (4 cards - 5% spawn rate)
| ID | Name | Strength | Loot Coins | Loot Gems | Rarity |
|---|---|---|---|---|---|
| `phoenix` | Eternal Phoenix | 11 | 2 | 6 | legendary |
| `hydra` | Nine-Headed Hydra | 12 | 2 | 7 | legendary |
| `kraken` | Abyssal Kraken | 13 | 2 | 8 | legendary |
| `leviathan` | Primordial Leviathan | 14 | 3 | 10 | legendary |

### Mythic Monster (1 card - 0.5% spawn rate)
| ID | Name | Strength | Loot Coins | Loot Gems | Rarity | Special |
|---|---|---|---|---|---|---|
| `ghost_king` | 👻 Ghost of the King | 15 | 4 | 15 | mythic | Cannot be challenged, Royal vengeance bonus |

---

## Powers Reference

### Active Powers (22+ powers)
| Power Name | Cost (Gems) | Target | Description |
|---|---|---|---|
| `change_monster` | 0 | monster | Swap one active monster with top card of monster deck |
| `fight_monster` | 1 | monster | Grants a free fight against a monster without normal restrictions |
| `fight_player_with_discarded_card` | 0 | other_player | Initiate strength duel using a hero from discard pile |
| `add_strength_to_challenger` | 2 | session | Add +1 strength during same-turn fight_player_with_discarded_card |
| `shuffle_heroes_deck` | 0 | session | Return all non-burned discarded heroes to deck and shuffle |
| `change_hero` | 2 | self | Draw a new hero card and discard your current one |
| `exchange_gems_with_player` | 0 | other_player | Swap your gems with those of a target player |
| `exchange_2coins_for_life` | 0 | self | Spend 2 coins to gain 1 life |
| `gain_2gems` | 0 | self | Instantly gain 2 gems |
| `gain_life` | 6 | self | Instantly gain 1 life |
| `see_player_card` | 0 | other_player | Secretly view another player's current hero card |
| `remove_coin` | 3 | other_player | Force target player to lose 1 coin (removed from game) |
| `shoot_player` | 0 | other_player | Guess target's hero; if correct, they lose 1 life |
| `all_in` | 69 | self | If last shoot_player correct: double gems; if wrong: lose all gems |
| `steal_gem` | 0 | other_player | Steal one gem from a target player |
| `parry_this_you_f_casual` | 6 | other_player | Reduce target player's life by 1 |
| `steal_coin` | 5 | other_player | Steal 1 coin from a target player |
| `mimic_power` | 6 | other_player | Copy and execute order-0 active power of target's hero |
| `mimic_hero` | 0 | other_player | Copy another player's hero abilities temporarily |
| `dual_attack` | 6 | session | After first monster fight, immediately fight next active monster |
| `blind_draw` | 0 | other_player | Draw a card blindly from the deck |
| `force_transform` | 4 | other_player | Force target to discard hero and draw new one |
| `execution` | 6 | session | Burn (permanently remove) a hero card from the game |
| `command_the_dead` | 4 | monster | Use discarded hero to fight monster and gain loot |

### Passive Powers (4 powers)
| Power Name | Trigger | Description |
|---|---|---|
| `keep_gems` | steal_attempt | Cancels steal attempts against you |
| `teamwork` | gems_gained | When you gain gems, duplicate the amount (double gems) |
| `alternate_strength` | turn_start | Strength alternates between 5 (human) and 12 (wolf) each turn |
| `tribute` | gems_gained | When another player gains gems, half are given to you instead |
| `royal_immunity` | always | Cannot be targeted by player attacks; gets no loot from direct monster fights |

---

## Special Mechanics

### Ghost of the King Special Rules
- **Cannot be challenged** when fought (no challenge window)
- **Mad Prince/Queen vengeance bonus:** +2 coins, +5 gems when they defeat it
- **Mad Queen command bonus:** +1 coin, +3 gems when commanding dead vs Ghost King
- **Extremely rare:** 0.5% spawn rate
- **Highest strength:** 15 (only Mad Prince at 13 or Wolf-form Werewolf at 12 can realistically defeat)

### Mad Queen Special Rules
- **Royal Immunity:** Cannot be targeted by player attacks like `parry_this_you_f_casual`
- **No direct loot:** Gets no loot from personal monster fights
- **Command the Dead:** Can use discarded heroes to fight monsters and claim loot
- **Unique playstyle:** Must rely on necromancy for resource generation

### Werewolf Transformation
- **Human form:** Strength 5 (vulnerable)
- **Wolf form:** Strength 12 (can fight legendary monsters)
- **Alternates each turn** via `alternate_strength` passive power
- **Dual nature:** Weakest and strongest hero depending on form

---

## Game Balance Notes

### Resource Economy
- **Coins:** Primary win condition (players + 1 to win)
- **Gems:** Power fuel (most powers cost 0-6 gems, `all_in` costs 69)
- **Life:** Survival resource (start with 3, lose 1 from failed fights/challenges)

### Monster Spawn Rates
- **Common:** 49.5% (easy early game targets)
- **Uncommon:** 30% (mid-game challenges)
- **Rare:** 15% (late game bosses)
- **Legendary:** 5% (epic encounters)
- **Mythic:** 0.5% (Ghost of the King - ultimate challenge)

### Power Balance
- **Free powers (0 gems):** Utility and basic actions
- **Low cost (1-3 gems):** Moderate impact abilities
- **High cost (4-6 gems):** Powerful game-changing effects
- **Ultra high cost (69 gems):** Ultimate risk/reward (`all_in`)

---

## Implementation Notes

### Database Schema Suggestions
```sql
-- Heroes table
heroes: id, name, strength, default_spawn_amount, power1, power2

-- Monsters table  
monsters: id, name, strength, loot_coins, loot_gems, rarity, spawn_rate

-- Powers table
powers: name, cost, target_type, is_passive, trigger_event, description

-- Game sessions should track:
-- - Active monsters (3 at a time)
-- - Player hero assignments
-- - Resource states (life, coins, gems)
-- - Discard/burn piles
-- - Challenge windows and timers
```

### Key Validation Rules
1. **Power cost validation:** Ensure player has enough gems before allowing power use
2. **Strength comparison:** Server-side validation for monster fights
3. **Challenge timing:** 10-second windows for fight/power moves (not demask)
4. **Royal immunity:** Mad Queen cannot be targeted by attack powers
5. **Ghost King special:** No challenge window, bonus loot for royals
6. **Werewolf transformation:** Automatic strength alternation each turn

---

*This reference covers all 15 heroes, 18 monsters, and 26+ powers in the current No Heroes No Lies meta.*