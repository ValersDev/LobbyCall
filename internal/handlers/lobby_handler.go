package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"valersbot/internal/enums"
	"valersbot/internal/errs"
	"valersbot/internal/models"
	"valersbot/internal/services"

	"github.com/bwmarrin/discordgo"
)

type LobbyHandler struct {
	lobby       *services.LobbyService
	lobbyRoleID string
}

func NewLobbyHandler(lobby *services.LobbyService, lobbyRoleID string) *LobbyHandler {
	return &LobbyHandler{
		lobby:       lobby,
		lobbyRoleID: lobbyRoleID,
	}
}

func (h *LobbyHandler) Register(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        enums.CommandPlanificar,
			Description: "Abre una votación para jugar",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        enums.OptionHora,
					Description: "Hora del plan (ej. 22:30)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        enums.OptionMensaje,
					Description: "Mensaje personalizado (opcional)",
					Required:    false,
					MaxLength:   200,
				},
			},
		},
		{
			Name:        enums.CommandCerrar,
			Description: "Cierra la votación abierta en este canal",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			return fmt.Errorf("creando comando /%s: %w", cmd.Name, err)
		}
		log.Printf("Comando /%s registrado correctamente.", cmd.Name)
	}

	return nil
}

func (h *LobbyHandler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		switch i.ApplicationCommandData().Name {
		case enums.CommandPlanificar:
			h.handlePlanificar(s, i)
		case enums.CommandCerrar:
			h.handleCerrar(s, i)
		}
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	}
}

func (h *LobbyHandler) handlePlanificar(s *discordgo.Session, i *discordgo.InteractionCreate) {
	horaRaw, ok := optionString(i.ApplicationCommandData().Options, enums.OptionHora)
	if !ok {
		respondEphemeral(s, i, "Falta la opción hora.")
		return
	}

	hora, err := services.ValidateHora(horaRaw)
	if err != nil {
		respondEphemeral(s, i, "Hora inválida. Usa el formato HH:MM (ej. 22:30).")
		return
	}

	mensaje, _ := optionString(i.ApplicationCommandData().Options, enums.OptionMensaje)
	mensaje = strings.TrimSpace(mensaje)

	if _, open := h.lobby.GetOpenPlanByChannel(i.ChannelID); open {
		respondEphemeral(s, i, "Ya hay una votación abierta en este canal. Usa `/cerrar` o el botón **Cerrar votación**.")
		return
	}

	plan := models.NewPlan("", i.ChannelID, hora, mensaje)
	content, components := h.formatPlanMessage(plan)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         content,
			Components:      components,
			AllowedMentions: h.roleAllowedMentions(),
		},
	})
	if err != nil {
		log.Printf("respondiendo /%s: %v", enums.CommandPlanificar, err)
		return
	}

	msg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		log.Printf("obteniendo mensaje de /%s: %v", enums.CommandPlanificar, err)
		return
	}

	if err := h.lobby.CreatePlan(msg.ID, i.ChannelID, hora, mensaje); err != nil {
		log.Printf("creando plan: %v", err)
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:    strPtr("No se pudo crear el plan. Inténtalo de nuevo."),
			Components: &[]discordgo.MessageComponent{},
		})
	}
}

func (h *LobbyHandler) handleCerrar(s *discordgo.Session, i *discordgo.InteractionCreate) {
	plan, err := h.lobby.ClosePlanByChannel(i.ChannelID)
	if err != nil {
		if errors.Is(err, errs.ErrNoOpenPlan) {
			respondEphemeral(s, i, "No hay una votación abierta en este canal.")
			return
		}
		log.Printf("cerrando plan: %v", err)
		respondEphemeral(s, i, "No se pudo cerrar la votación.")
		return
	}

	content, _ := h.formatPlanMessage(plan)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    plan.ChannelID,
		ID:         plan.ID,
		Content:    &content,
		Components: &[]discordgo.MessageComponent{},
	})
	if err != nil {
		log.Printf("editando mensaje cerrado: %v", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         h.closeAnnouncement(plan),
			AllowedMentions: h.roleAllowedMentions(),
		},
	})
	if err != nil {
		log.Printf("anunciando cierre: %v", err)
	}
}

func (h *LobbyHandler) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	choice := i.MessageComponentData().CustomID
	switch {
	case isVoteChoice(choice):
		h.handleVote(s, i, choice)
	case choice == enums.ClosePlan:
		h.handleCloseButton(s, i)
	}
}

func (h *LobbyHandler) handleVote(s *discordgo.Session, i *discordgo.InteractionCreate, choice string) {
	user := interactionUser(i)
	if user == nil {
		respondEphemeral(s, i, "No se pudo identificar al usuario.")
		return
	}

	plan, err := h.lobby.RegisterVote(i.Message.ID, models.Player{
		ID:       user.ID,
		Username: user.Username,
	}, choice)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlanClosed):
			respondEphemeral(s, i, "Esta votación ya está cerrada.")
		case errors.Is(err, errs.ErrPlanNotFound):
			respondEphemeral(s, i, "No se encontró esta votación.")
		default:
			log.Printf("registrando voto: %v", err)
			respondEphemeral(s, i, "No se pudo registrar el voto.")
		}
		return
	}

	content, components := h.formatPlanMessage(plan)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: components,
		},
	})
	if err != nil {
		log.Printf("actualizando mensaje de votos: %v", err)
	}
}

func (h *LobbyHandler) handleCloseButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	plan, err := h.lobby.ClosePlan(i.Message.ID)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPlanClosed):
			respondEphemeral(s, i, "Esta votación ya está cerrada.")
		case errors.Is(err, errs.ErrPlanNotFound):
			respondEphemeral(s, i, "No se encontró esta votación.")
		default:
			log.Printf("cerrando plan por botón: %v", err)
			respondEphemeral(s, i, "No se pudo cerrar la votación.")
		}
		return
	}

	content, _ := h.formatPlanMessage(plan)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		log.Printf("actualizando mensaje cerrado: %v", err)
		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content:         h.closeAnnouncement(plan),
		AllowedMentions: h.roleAllowedMentions(),
	})
	if err != nil {
		log.Printf("anunciando cierre por followup: %v", err)
	}
}

func (h *LobbyHandler) roleMention() string {
	if h.lobbyRoleID == "" {
		return ""
	}
	return fmt.Sprintf("<@&%s> ", h.lobbyRoleID)
}

func (h *LobbyHandler) roleAllowedMentions() *discordgo.MessageAllowedMentions {
	if h.lobbyRoleID == "" {
		return &discordgo.MessageAllowedMentions{}
	}
	return &discordgo.MessageAllowedMentions{
		Parse: []discordgo.AllowedMentionType{},
		Roles: []string{h.lobbyRoleID},
	}
}

func (h *LobbyHandler) closeAnnouncement(plan *models.Plan) string {
	return h.roleMention() + formatWinnerLine(plan)
}

func (h *LobbyHandler) formatPlanMessage(plan *models.Plan) (string, []discordgo.MessageComponent) {
	var b strings.Builder
	if mention := h.roleMention(); mention != "" {
		b.WriteString(mention)
		b.WriteString("\n")
	}

	if plan.Closed {
		b.WriteString(fmt.Sprintf("**Votación cerrada — plan para las %s**\n", plan.Time))
		if plan.Note != "" {
			b.WriteString(plan.Note)
			b.WriteString("\n")
		}
		b.WriteString(formatWinnerLine(plan))
		b.WriteString("\n\n")
	} else {
		b.WriteString(fmt.Sprintf("**¡Plan propuesto para las %s!**\n", plan.Time))
		if plan.Note != "" {
			b.WriteString(plan.Note)
			b.WriteString("\n")
		}
		b.WriteString("¿A qué jugamos?\n")
		if leading := formatLeadingLine(plan); leading != "" {
			b.WriteString(leading)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf(
		"**Flex (%d):** %s\n"+
			"**5v5 (%d):** %s\n"+
			"**ARAM (%d):** %s\n"+
			"**Chill (%d):** %s",
		len(plan.VotesFlex), formatVoters(plan.VotesFlex),
		len(plan.Votes5v5), formatVoters(plan.Votes5v5),
		len(plan.VotesAram), formatVoters(plan.VotesAram),
		len(plan.VotesChill), formatVoters(plan.VotesChill),
	))

	if plan.Closed {
		return b.String(), nil
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    fmt.Sprintf("LoL Flex (%d)", len(plan.VotesFlex)),
					Style:    discordgo.PrimaryButton,
					CustomID: enums.VoteFlex,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("LoL 5v5 (%d)", len(plan.Votes5v5)),
					Style:    discordgo.SuccessButton,
					CustomID: enums.Vote5v5,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("LoL ARAM (%d)", len(plan.VotesAram)),
					Style:    discordgo.SuccessButton,
					CustomID: enums.VoteARAM,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Otro / Chill (%d)", len(plan.VotesChill)),
					Style:    discordgo.SecondaryButton,
					CustomID: enums.VoteChill,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Cerrar votación",
					Style:    discordgo.DangerButton,
					CustomID: enums.ClosePlan,
				},
			},
		},
	}

	return b.String(), components
}

func formatLeadingLine(plan *models.Plan) string {
	labels, count := plan.Winners()
	if count == 0 {
		return ""
	}
	if len(labels) == 1 {
		return fmt.Sprintf("**Va ganando:** %s (%d)", labels[0], count)
	}
	return fmt.Sprintf("**Empate:** %s (%d)", strings.Join(labels, " / "), count)
}

func formatWinnerLine(plan *models.Plan) string {
	labels, count := plan.Winners()
	switch {
	case count == 0:
		return fmt.Sprintf("Sin votos. El plan era para las **%s**.", plan.Time)
	case len(labels) == 1:
		return fmt.Sprintf("¡A las **%s** jugamos **%s**! (%d voto%s)", plan.Time, labels[0], count, plural(count))
	default:
		return fmt.Sprintf("Empate a las **%s** entre **%s** (%d voto%s).", plan.Time, strings.Join(labels, " / "), count, plural(count))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func formatVoters(votes map[string]models.Player) string {
	if len(votes) == 0 {
		return "—"
	}

	names := make([]string, 0, len(votes))
	for _, p := range votes {
		names = append(names, p.Username)
	}
	return strings.Join(names, ", ")
}

func isVoteChoice(choice string) bool {
	switch choice {
	case enums.VoteFlex, enums.Vote5v5, enums.VoteARAM, enums.VoteChill:
		return true
	default:
		return false
	}
}

func interactionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
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

func strPtr(s string) *string {
	return &s
}
