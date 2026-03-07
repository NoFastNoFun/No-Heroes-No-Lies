package powers

import "context"

func RegisterAll(reg *Registry) {
	RegisterPassiveListeners()
	RegisterActiveResource(reg)
	RegisterActiveDeck(reg)
	RegisterActiveCombat(reg)
	RegisterActiveMimic(reg)

	noop := func(context.Context, ExecutionContext, PowerRunner) error { return nil }
	reg.Register("keep_gems", PowerDef{Name: "keep_gems", CostGems: 0, TargetType: TargetSelf, IsPassive: true}, noop)
	reg.Register("teamwork", PowerDef{Name: "teamwork", CostGems: 0, TargetType: TargetSelf, IsPassive: true}, noop)
	reg.Register("alternate_strength", PowerDef{Name: "alternate_strength", CostGems: 0, TargetType: TargetSelf, IsPassive: true}, noop)
	reg.Register("tribute", PowerDef{Name: "tribute", CostGems: 0, TargetType: TargetSelf, IsPassive: true}, noop)
	reg.Register("royal_immunity", PowerDef{Name: "royal_immunity", CostGems: 0, TargetType: TargetSelf, IsPassive: true}, noop)
}
