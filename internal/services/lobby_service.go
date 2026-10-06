package services

import (
	"fmt"
	"regexp"
	"strconv"

	"lobbycall/internal/enums"
	"lobbycall/internal/errs"
	"lobbycall/internal/models"
	"lobbycall/internal/storage"
)

var horaPattern = regexp.MustCompile(`^([01]?\d|2[0-3]):([0-5]\d)$`)

type LobbyService struct {
	store *storage.MemoryStore
}

func NewLobbyService(store *storage.MemoryStore) *LobbyService {
	return &LobbyService{store: store}
}

func ValidateHora(raw string) (string, error) {
	matches := horaPattern.FindStringSubmatch(raw)
	if matches == nil {
		return "", errs.ErrInvalidHora
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])
	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}

func (s *LobbyService) CreatePlan(messageID, channelID, planTime, note string) error {
	return s.store.CreateOpenPlan(models.NewPlan(messageID, channelID, planTime, note))
}

func (s *LobbyService) RegisterVote(messageID string, player models.Player, choice string) (*models.Plan, error) {
	return s.store.UpdatePlan(messageID, func(plan *models.Plan) error {
		if plan.Closed {
			return errs.ErrPlanClosed
		}

		delete(plan.VotesFlex, player.ID)
		delete(plan.Votes5v5, player.ID)
		delete(plan.VotesAram, player.ID)
		delete(plan.VotesChill, player.ID)

		switch choice {
		case enums.VoteFlex:
			plan.VotesFlex[player.ID] = player
		case enums.Vote5v5:
			plan.Votes5v5[player.ID] = player
		case enums.VoteARAM:
			plan.VotesAram[player.ID] = player
		case enums.VoteChill:
			plan.VotesChill[player.ID] = player
		default:
			return fmt.Errorf("unknown vote choice: %s", choice)
		}

		return nil
	})
}

func (s *LobbyService) ClosePlan(messageID string) (*models.Plan, error) {
	return s.store.UpdatePlan(messageID, func(plan *models.Plan) error {
		if plan.Closed {
			return errs.ErrPlanClosed
		}
		plan.Closed = true
		return nil
	})
}

func (s *LobbyService) ClosePlanByChannel(channelID string) (*models.Plan, error) {
	plan, ok := s.store.GetOpenPlanByChannel(channelID)
	if !ok {
		return nil, errs.ErrNoOpenPlan
	}
	return s.ClosePlan(plan.ID)
}

func (s *LobbyService) GetOpenPlanByChannel(channelID string) (*models.Plan, bool) {
	return s.store.GetOpenPlanByChannel(channelID)
}
