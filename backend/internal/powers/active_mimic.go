package powers

import "context"

func RegisterActiveMimic(reg *Registry) {
	reg.Register("mimic_power", PowerDef{Name: "mimic_power", CostGems: 6, TargetType: TargetOtherPlayer}, execMimicPower)
	reg.Register("mimic_hero", PowerDef{Name: "mimic_hero", CostGems: 0, TargetType: TargetOtherPlayer}, execMimicHero)
}

func execMimicPower(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	hero, err := r.GetHeroByID(ctx, *target.HeroID)
	if err != nil {
		return err
	}
	powerName := hero.Power1
	if powerName == "mimic_power" || powerName == "mimic_hero" {
		return errBadRequest
	}
	subEC := ExecutionContext{
		GameID:       ec.GameID,
		ActorID:      ec.ActorID,
		TargetPlayer: ec.TargetPlayer,
		MonsterSlot:  ec.MonsterSlot,
		Payload:      ec.Payload,
	}
	return r.ExecutePower(ctx, powerName, subEC)
}

func execMimicHero(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	heroID := *target.HeroID
	return r.SetMimicHero(ctx, ec.GameID, &ec.ActorID, &heroID)
}
