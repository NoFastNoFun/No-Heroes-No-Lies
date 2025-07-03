-- Initial card data for No Heroes No Lies
CREATE TABLE IF NOT EXISTS cards (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    strength INT NOT NULL,
    default_spawn INT,
    rarity VARCHAR,
    power1 VARCHAR,
    power2 VARCHAR,
    notes VARCHAR
);

INSERT INTO
    cards (
        id,
        name,
        strength,
        default_spawn,
        rarity,
        power1,
        power2,
        notes
    )
VALUES (
        'troll',
        'Troll',
        7,
        2,
        'Uncommon',
        'fight_monster',
        'change_monster',
        'Combat specialist'
    ),
    (
        'esprit',
        'Esprit Sylvestre',
        6,
        2,
        'Uncommon',
        'fight_player_with_discarded_card',
        'add_strength_to_challenger',
        'PvP fighter'
    ),
    (
        'barde',
        'Barde',
        2,
        1,
        'Rare',
        'shuffle_heroes_deck',
        'change_hero',
        'Deck manipulation'
    ),
    (
        'elf',
        'Elf',
        8,
        2,
        'Uncommon',
        'exchange_gems_with_player',
        'exchange_2coins_for_life',
        'Resource trading'
    ),
    (
        'nain',
        'Nain',
        5,
        2,
        'Uncommon',
        'gain_2gems',
        'gain_life',
        'Resource generation'
    ),
    (
        'mage',
        'Mage',
        9,
        2,
        'Uncommon',
        'see_player_card',
        'remove_coin',
        'Information & disruption'
    ),
    (
        'exorciste',
        'Exorciste',
        3,
        4,
        'Common',
        'shoot_player',
        'all_in',
        'High risk/reward'
    ),
    (
        'gobelin',
        'Gobelin',
        4,
        2,
        'Uncommon',
        'steal_gem',
        'steal_gem',
        'Gem thief (dual steal)'
    ),
    (
        'geant',
        'Géant',
        10,
        1,
        'Rare',
        'keep_gems',
        'parry_this_you_f_casual',
        'Tank with damage'
    ),
    (
        'chevalier',
        'Chevalier',
        11,
        1,
        'Rare',
        'teamwork',
        'steal_coin',
        'Elite knight'
    ),
    (
        'shapeshifter',
        'Shapeshifter',
        1,
        1,
        'Rare',
        'mimic_hero',
        'mimic_power',
        'Copycat abilities'
    ),
    (
        'werewolf',
        'Werewolf',
        0,
        2,
        'Uncommon',
        'alternate_strength',
        'dual_attack',
        'Alternates 5/12'
    ),
    (
        'witch',
        'Witch',
        8,
        1,
        'Rare',
        'blind_draw',
        'force_transform',
        'Transformation magic'
    ),
    (
        'prince',
        'Mad Prince',
        13,
        1,
        'Rare',
        'tribute',
        'execution',
        'Strongest hero'
    ),
    (
        'queen',
        'Mad Queen',
        9,
        1,
        'Rare',
        'royal_immunity',
        'command_the_dead',
        'Immune to attacks'
    ) ON CONFLICT (id) DO NOTHING;