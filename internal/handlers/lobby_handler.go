package handlers

import (
	"fmt"
	"log"

	"valersbot/internal/enums"

	"github.com/bwmarrin/discordgo"
)

type LobbyHandler struct{}

func NewLobbyHandler() *LobbyHandler {
	return &LobbyHandler{}
}

func (h *LobbyHandler) Register(s *discordgo.Session, guildID string) error {
	comando := &discordgo.ApplicationCommand{
		Name:        enums.CommandPlanificar,
		Description: "Abre una votación para jugar",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        enums.OptionHora,
				Description: "Hora del plan (ej. 22:30)",
				Required:    true,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, comando)
	if err != nil {
		return fmt.Errorf("creando comando /%s: %w", enums.CommandPlanificar, err)
	}

	log.Printf("Comando /%s registrado correctamente.", enums.CommandPlanificar)
	return nil
}

func (h *LobbyHandler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if i.ApplicationCommandData().Name == enums.CommandPlanificar {
			h.handlePlanificar(s, i)
		}
	case discordgo.InteractionMessageComponent:
		h.handleVote(s, i)
	}
}

func (h *LobbyHandler) handlePlanificar(s *discordgo.Session, i *discordgo.InteractionCreate) {
	hora, ok := optionString(i.ApplicationCommandData().Options, enums.OptionHora)
	if !ok {
		respondEphemeral(s, i, "Falta la opción hora.")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**¡Plan propuesto para las %s!**\n¿A qué jugamos?", hora),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{Label: "LoL Flex", Style: discordgo.PrimaryButton, CustomID: enums.VoteFlex},
						discordgo.Button{Label: "LoL 5v5", Style: discordgo.SuccessButton, CustomID: enums.Vote5v5},
						discordgo.Button{Label: "LoL ARAM", Style: discordgo.SuccessButton, CustomID: enums.VoteARAM},
						discordgo.Button{Label: "Otro / Chill", Style: discordgo.SecondaryButton, CustomID: enums.VoteChill},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("respondiendo /%s: %v", enums.CommandPlanificar, err)
	}
}

func (h *LobbyHandler) handleVote(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: tally votes for enums.VoteFlex / Vote5v5 / VoteARAM / VoteChill
}

func optionString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) (string, bool) {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue(), true
		}
	}
	return "", false
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("respondiendo ephemeral: %v", err)
	}
}
