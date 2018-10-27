package requests

import (
	"github.com/go-errors/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/clagraff/pitch/comms/responses"
	"github.com/clagraff/pitch/entities"
	"github.com/clagraff/pitch/entities/items"
	"github.com/clagraff/pitch/entities/objects"
	"github.com/clagraff/pitch/logging"
	"github.com/clagraff/pitch/utils"
)

// calcLuckyCritical will calculate whether a critical hit has occurred based on
// a roll against the Luck attribute of the specified actor.
func calcLuckyCritical(actor objects.Entity) bool {
	// If Luck modifier + 5 is greater than the roll, a crit has occurred.
	//		0% chance at lowest Luck level, Luck <= 1.
	//		15% change at highest Luck leve, Luck >= 30

	luck := actor.Attributes.Luck.Modifier() + 5
	luckRoll := utils.Roll(1, 100)

	return luck >= luckRoll
}

// calcAttackRoll will determine the attacker's hit roll, and whether a crit
// has occurred.
// A crit occurs either due to a good Luck roll, or if their attack roll is a
// perfect `20`.
func calcAttackRoll(attacker objects.Entity) (int, bool) {
	critical := calcLuckyCritical(attacker)

	attackRoll := utils.Roll(1, 20)
	attackModifier := attacker.Attributes.Strength.Modifier()

	if attackRoll == 20 {
		critical = true
	}

	return attackRoll + attackModifier, critical
}

// calcDamageAmount determines the amount of damage an attacker does with their
// current primary item.
// Takes into account adding the Strength modifier for melee items, and
// Dexterity modifier for range items.
func calcDamageAmount(world entities.World, attacker objects.Entity) (int, items.DamageType) {
	damageType := items.MeleeDamage
	damageRoll := utils.Roll(1, 4) // Unarmed attack is default. 1d4+0

	if item, ok := world.Items.FromID(attacker.Equipment.PrimaryItemID); ok {
		damageRoll = item.Damage.Calculate()
		damageType = item.Damage.Type

		if damageType == items.MeleeDamage {
			damageRoll = damageRoll + attacker.Attributes.Strength.Modifier()
		} else if damageType == items.RangeDamage {
			damageRoll = damageRoll + attacker.Attributes.Dexterity.Modifier()
		}
	}

	if damageRoll < 0 {
		damageRoll = 0
	}

	return damageRoll, damageType
}

func isHit(isCritical bool, armorClass int, attackRoll int) bool {
	return isCritical || armorClass <= attackRoll
}

// MeleeAttackRequest represents an attack request by the Attacker against the
// target, using a melee weapon.
type MeleeAttackRequest struct {
	AttackerID uuid.UUID `json:"attacker_id"`
	TargetID   uuid.UUID `json:"target_id"`
}

// Execute performs the melee request.
func (req MeleeAttackRequest) Execute(world entities.World) (entities.World, responses.Response, error) {
	logger, closeLog := logging.Logger("requests.melee_attack_request")
	defer closeLog()

	attacker, ok := world.Objects.FromID(req.AttackerID)
	if !ok {
		return world, nil, errors.Errorf("attacker could not be found")
	}

	target, ok := world.Objects.FromID(req.TargetID)
	if !ok {
		return world, nil, errors.Errorf("target could not be found")
	}

	attackRoll, critical := calcAttackRoll(attacker)

	ac := 10 + target.Attributes.Dexterity.Modifier()

	// To hit, the attack roll + modifier must be greater than target AC.
	// Always hits if is critical.
	if !isHit(critical, ac, attackRoll) {
		resp := responses.MeleeAttackResponse{
			AttackerID:      req.AttackerID,
			TargetID:        req.TargetID,
			DidHit:          false,
			Damage:          0,
			HealthRemaining: target.Health,
		}
		return world, resp, nil
	}

	damageAmount, damateType := calcDamageAmount(world, attacker)
	if damateType != items.MeleeDamage {
		return world, nil, errors.Errorf("tried to melee attack with non-melee weapon")
	}

	// When a critical occurs, the attack always hits and the damage is doubled.
	// Double damage during crits.
	if critical {
		damageAmount *= 2
	}

	armor := 0
	if item, ok := world.Items.FromID(target.Equipment.ChestID); ok {
		armor = armor + item.Armor.MeleeReduction
	}

	if armor < 0 {
		armor = 0
	}

	totalDamage := armor - damageAmount
	if totalDamage < 0 {
		totalDamage = 0
	}

	logger.Printf(
		"%s melee attacked %s: %d damage\n",
		attacker.ID.String(),
		target.ID.String(),
		totalDamage,
	)

	target.Health = target.Health - totalDamage
	if target.Health <= 0 {
		target.Health = 0
		world.Objects = world.Objects.MustRemove(target)

		logger.Printf(
			"%s has been killed and removed",
			target.ID.String(),
		)
	} else {
		world.Objects = world.Objects.MustUpdate(target)
	}

	resp := responses.MeleeAttackResponse{
		AttackerID:      req.AttackerID,
		TargetID:        req.TargetID,
		DidHit:          true,
		Damage:          totalDamage,
		HealthRemaining: target.Health,
	}
	return world, resp, nil
}

// RangeAttackRequest represents an attack request by the Attacker against the
// target, using a range weapon.
type RangeAttackRequest struct {
	AttackerID uuid.UUID `json:"attacker_id"`
	TargetID   uuid.UUID `json:"target_id"`
}

// Execute performs the range attack request.
func (req RangeAttackRequest) Execute(world entities.World) (entities.World, responses.Response, error) {
	logger, closeLog := logging.Logger("requests.range_attack_request")
	defer closeLog()

	attacker, ok := world.Objects.FromID(req.AttackerID)
	if !ok {
		return world, nil, errors.Errorf("attacker could not be found")
	}

	target, ok := world.Objects.FromID(req.TargetID)
	if !ok {
		return world, nil, errors.Errorf("target could not be found")
	}

	attackRoll, critical := calcAttackRoll(attacker)

	ac := 10 + target.Attributes.Dexterity.Modifier()

	// To hit, the attack roll + modifier must be greater than target AC.
	// Always hits if is critical.
	if !isHit(critical, ac, attackRoll) {
		resp := responses.RangeAttackResponse{
			AttackerID:      req.AttackerID,
			TargetID:        req.TargetID,
			DidHit:          false,
			Damage:          0,
			HealthRemaining: target.Health,
		}
		return world, resp, nil
	}

	damageAmount, damageType := calcDamageAmount(world, attacker)
	if damageType != items.RangeDamage {
		return world, nil, errors.Errorf("tried to range attack with non-range weapon")
	}

	// When a critical occurs, the attack always hits and the damage is doubled.
	// Double damage during crits.
	if critical {
		damageAmount *= 2
	}

	armor := 0
	if item, ok := world.Items.FromID(target.Equipment.ChestID); ok {
		armor = armor + item.Armor.RangeReduction
	}

	if armor < 0 {
		armor = 0
	}

	totalDamage := armor - damageAmount
	if totalDamage < 0 {
		totalDamage = 0
	}

	logger.Printf(
		"%s range attacked %s: %d damage\n",
		attacker.ID.String(),
		target.ID.String(),
		totalDamage,
	)

	target.Health = target.Health - totalDamage
	if target.Health <= 0 {
		target.Health = 0
		world.Objects = world.Objects.MustRemove(target)

		logger.Printf(
			"%s has been killed and removed",
			target.ID.String(),
		)
	} else {
		world.Objects = world.Objects.MustUpdate(target)
	}

	resp := responses.RangeAttackResponse{
		AttackerID:      req.AttackerID,
		TargetID:        req.TargetID,
		DidHit:          true,
		Damage:          totalDamage,
		HealthRemaining: target.Health,
	}
	return world, resp, nil
}
