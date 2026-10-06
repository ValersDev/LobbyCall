package models

type Player struct {
	ID       string
	Username string
}

type Plan struct {
	ID         string
	ChannelID  string
	Time       string
	Note       string
	Closed     bool
	VotesFlex  map[string]Player
	Votes5v5   map[string]Player
	VotesAram  map[string]Player
	VotesChill map[string]Player
}

func NewPlan(messageID, channelID, planTime, note string) *Plan {
	return &Plan{
		ID:         messageID,
		ChannelID:  channelID,
		Time:       planTime,
		Note:       note,
		VotesFlex:  make(map[string]Player),
		Votes5v5:   make(map[string]Player),
		VotesAram:  make(map[string]Player),
		VotesChill: make(map[string]Player),
	}
}

// Winners returns the leading option labels and their vote count.
// If there are no votes, labels is empty and count is 0.
func (p *Plan) Winners() (labels []string, count int) {
	options := []struct {
		label string
		n     int
	}{
		{"Flex", len(p.VotesFlex)},
		{"5v5", len(p.Votes5v5)},
		{"ARAM", len(p.VotesAram)},
		{"Chill", len(p.VotesChill)},
	}

	for _, opt := range options {
		if opt.n > count {
			count = opt.n
			labels = []string{opt.label}
			continue
		}
		if opt.n > 0 && opt.n == count {
			labels = append(labels, opt.label)
		}
	}

	if count == 0 {
		return nil, 0
	}
	return labels, count
}

func (p *Plan) TotalVotes() int {
	return len(p.VotesFlex) + len(p.Votes5v5) + len(p.VotesAram) + len(p.VotesChill)
}
