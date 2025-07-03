-- Initial powers data for No Heroes No Lies
CREATE TABLE IF NOT EXISTS powers (
    name VARCHAR PRIMARY KEY,
    cost INT,
    target VARCHAR,
    description VARCHAR,
    is_passive BOOLEAN DEFAULT FALSE,
    trigger VARCHAR
);

INSERT INTO
    powers (
        name,
        cost,
        target,
        description,
        is_passive,
        trigger
    )
VALUES (
        'change_monster',
        0,
        'monster',
        'Swap one active monster with top card of monster deck',
        FALSE,
        NULL
    ),
    (
        'fight_monster',
        1,
        'monster',
        'Grants a free fight against a monster without normal restrictions',
        FALSE,
        NULL
    ),
    (
        'fight_player_with_discarded_card',
        0,
        'other_player',
        'Initiate strength duel using a hero from discard pile',
        FALSE,
        NULL
    ),
    (
        'add_strength_to_challenger',
        2,
        'session',
        'Add +1 strength during same-turn fight_player_with_discarded_card',
        FALSE,
        NULL
    ),
    (
        'shuffle_heroes_deck',
        0,
        'session',
        'Return all non-burned discarded heroes to deck and shuffle',
        FALSE,
        NULL
    ),
    (
        'change_hero',
        2,
        'self',
        'Draw a new hero card and discard your current one',
        FALSE,
        NULL
    ),
    (
        'exchange_gems_with_player',
        0,
        'other_player',
        'Swap your gems with those of a target player',
        FALSE,
        NULL
    ),
    (
        'exchange_2coins_for_life',
        0,
        'self',
        'Spend 2 coins to gain 1 life',
        FALSE,
        NULL
    ),
    (
        'gain_2gems',
        0,
        'self',
        'Instantly gain 2 gems',
        FALSE,
        NULL
    ),
    (
        'gain_life',
        6,
        'self',
        'Instantly gain 1 life',
        FALSE,
        NULL
    ),
    (
        'see_player_card',
        0,
        'other_player',
        'Secretly view another player''s current hero card',
        FALSE,
        NULL
    ),
    (
        'remove_coin',
        3,
        'other_player',
        'Force target player to lose 1 coin (removed from game)',
        FALSE,
        NULL
    ),
    (
        'shoot_player',
        0,
        'other_player',
        'Guess target''s hero; if correct, they lose 1 life',
        FALSE,
        NULL
    ),
    (
        'all_in',
        69,
        'self',
        'If last shoot_player correct: double gems; if wrong: lose all gems',
        FALSE,
        NULL
    ),
    (
        'steal_gem',
        0,
        'other_player',
        'Steal one gem from a target player',
        FALSE,
        NULL
    ),
    (
        'parry_this_you_f_casual',
        6,
        'other_player',
        'Reduce target player''s life by 1',
        FALSE,
        NULL
    ),
    (
        'steal_coin',
        5,
        'other_player',
        'Steal 1 coin from a target player',
        FALSE,
        NULL
    ),
    (
        'mimic_power',
        6,
        'other_player',
        'Copy and execute order-0 active power of target''s hero',
        FALSE,
        NULL
    ),
    (
        'mimic_hero',
        0,
        'other_player',
        'Copy another player''s hero abilities temporarily',
        FALSE,
        NULL
    ),
    (
        'dual_attack',
        6,
        'session',
        'After first monster fight, immediately fight next active monster',
        FALSE,
        NULL
    ),
    (
        'blind_draw',
        0,
        'other_player',
        'Draw a card blindly from the deck',
        FALSE,
        NULL
    ),
    (
        'force_transform',
        4,
        'other_player',
        'Force target to discard hero and draw new one',
        FALSE,
        NULL
    ),
    (
        'execution',
        6,
        'session',
        'Burn (permanently remove) a hero card from the game',
        FALSE,
        NULL
    ),
    (
        'command_the_dead',
        4,
        'monster',
        'Use discarded hero to fight monster and gain loot',
        FALSE,
        NULL
    ),
    (
        'keep_gems',
        0,
        NULL,
        'Cancels steal attempts against you',
        TRUE,
        'steal_attempt'
    ),
    (
        'teamwork',
        0,
        NULL,
        'When you gain gems, duplicate the amount (double gems)',
        TRUE,
        'gems_gained'
    ),
    (
        'alternate_strength',
        0,
        NULL,
        'Strength alternates between 5 (human) and 12 (wolf) each turn',
        TRUE,
        'turn_start'
    ),
    (
        'tribute',
        0,
        NULL,
        'When another player gains gems, half are given to you instead',
        TRUE,
        'gems_gained'
    ),
    (
        'royal_immunity',
        0,
        NULL,
        'Cannot be targeted by player attacks; gets no loot from direct monster fights',
        TRUE,
        'always'
    ) ON CONFLICT (name) DO NOTHING;