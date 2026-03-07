package powers

import (
	"context"
	"errors"
)

func RegisterActiveResource(reg *Registry) {
	reg.Register("gain_2gems", PowerDef{Name: "gain_2gems", CostGems: 0, TargetType: TargetSelf}, execGain2Gems)
	reg.Register("gain_life", PowerDef{Name: "gain_life", CostGems: 6, TargetType: TargetSelf}, execGainLife)
	reg.Register("exchange_2coins_for_life", PowerDef{Name: "exchange_2coins_for_life", CostGems: 0, TargetType: TargetSelf}, execExchange2CoinsForLife)
	reg.Register("exchange_gems_with_player", PowerDef{Name: "exchange_gems_with_player", CostGems: 0, TargetType: TargetOtherPlayer}, execExchangeGemsWithPlayer)
	reg.Register("remove_coin", PowerDef{Name: "remove_coin", CostGems: 3, TargetType: TargetOtherPlayer}, execRemoveCoin)
	reg.Register("steal_coin", PowerDef{Name: "steal_coin", CostGems: 5, TargetType: TargetOtherPlayer}, execStealCoin)
	reg.Register("steal_gem", PowerDef{Name: "steal_gem", CostGems: 0, TargetType: TargetOtherPlayer}, execStealGem)
	reg.Register("parry_this_you_f_casual", PowerDef{Name: "parry_this_you_f_casual", CostGems: 6, TargetType: TargetOtherPlayer}, execParryThisYouFCasual)
}

func execGain2Gems(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if err := r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, 2); err != nil {
		return err
	}
	r.EmitEvent(ctx, &Event{
		Kind:         EventKindGemsGained,
		GemsGameID:   ec.GameID,
		GemsPlayerID: ec.ActorID,
		GemsAmount:   2,
		GemsIsBonus:  false,
	})
	return nil
}

func execGainLife(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	return r.AddPlayerLife(ctx, ec.GameID, ec.ActorID, 1)
}

func execExchange2CoinsForLife(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	actor, err := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	if actor.Coins < 2 {
		return errInsufficientCoins
	}
	if err := r.AddPlayerCoins(ctx, ec.GameID, ec.ActorID, -2); err != nil {
		return err
	}
	return r.AddPlayerLife(ctx, ec.GameID, ec.ActorID, 1)
}

func execExchangeGemsWithPlayer(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	actor, err := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil {
		return err
	}
	actorGems := actor.Gems
	targetGems := target.Gems
	if err := r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, targetGems-actorGems); err != nil {
		return err
	}
	return r.AddPlayerGems(ctx, ec.GameID, *ec.TargetPlayer, actorGems-targetGems)
}

func execRemoveCoin(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	return r.AddPlayerCoins(ctx, ec.GameID, *ec.TargetPlayer, -1)
}

func execStealCoin(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	if HasRoyalImmunity(r, ctx, ec.GameID, *ec.TargetPlayer) {
		return nil
	}
	if err := r.AddPlayerCoins(ctx, ec.GameID, *ec.TargetPlayer, -1); err != nil {
		return err
	}
	return r.AddPlayerCoins(ctx, ec.GameID, ec.ActorID, 1)
}

func execStealGem(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	if HasRoyalImmunity(r, ctx, ec.GameID, *ec.TargetPlayer) {
		return nil
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.Gems < 1 {
		return errBadRequest
	}
	ev := &Event{
		Kind:          EventKindStealAttempt,
		StealGameID:   ec.GameID,
		StealerID:     ec.ActorID,
		TargetID:      *ec.TargetPlayer,
		StealResource: "gem",
	}
	r.EmitEvent(ctx, ev)
	if ev.Cancelled {
		return nil
	}
	if err := r.AddPlayerGems(ctx, ec.GameID, *ec.TargetPlayer, -1); err != nil {
		return err
	}
	return r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, 1)
}

func execParryThisYouFCasual(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	if HasRoyalImmunity(r, ctx, ec.GameID, *ec.TargetPlayer) {
		return nil
	}
	return r.AddPlayerLife(ctx, ec.GameID, *ec.TargetPlayer, -1)
}

var (
	errBadRequest        = errors.New("bad request")
	errInsufficientCoins = errors.New("insufficient coins")
)
