Main gameplay gaps / mismatches
Game start / basic state
Players are never actually added to GameState.Players, no TurnOrder or CurrentTurn is set, and no initial Life / Coins / Gems / CurrentHero are assigned in the backend, so turn-taking and starting positions are effectively undefined in code.
Passives not wired
Werewolf / alternate_strength: passive exists in data, but there is no trigger or logic to alternate between 5/12 strength each turn.
Mad Queen / royal_immunity: immunity is defined but never checked:
She can be targeted by attack powers (shoot, parry, steal, etc.) in code.
She still receives normal loot from her own monster fights, contrary to “no direct loot”.
Gem economy & triggers
Powers that give gems directly (e.g. gain_2gems) just increment gems and don’t fire the gems_gained event, so teamwork (double gems) and tribute (tax half to Prince) don’t react to them even though the design says they should trigger on gem gains in general.
Ghost of the King special rules
Treated like a normal monster:
Can be challenged (no “cannot be challenged” exception).
No extra loot or vengeance bonuses for Mad Prince / Mad Queen.
So its unique meta rules from the doc are not implemented.
Mimic power
mimic_power is explicitly “not yet implemented”; Shapeshifter’s second power does nothing.
Challenge system details
LastMove is meant to store coins/gems deltas to roll back gains on a successful challenge, but those deltas are never filled in, so resource rollback on catching a liar is effectively missing.