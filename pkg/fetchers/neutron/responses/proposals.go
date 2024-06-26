package responses

import (
	"main/pkg/types"
	"strconv"
	"time"

	"cosmossdk.io/math"
)

type ProposalWithID struct {
	ID       int      `json:"id"`
	Proposal Proposal `json:"proposal"`
}
type Proposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Expiration  struct {
		AtTime string `json:"at_time"`
	} `json:"expiration"`
	Status     string `json:"status"`
	TotalPower string `json:"total_power"`

	Votes struct {
		Yes     string `json:"yes"`
		No      string `json:"no"`
		Abstain string `json:"abstain"`
	} `json:"votes"`
}

type ProposalsResponse struct {
	Data struct {
		Proposals []ProposalWithID `json:"proposals"`
	} `json:"data"`
}

func (p ProposalsResponse) ToProposals() ([]types.Proposal, error) {
	proposals := make([]types.Proposal, len(p.Data.Proposals))

	for index, proposal := range p.Data.Proposals {
		proposalParsed, err := proposal.ToProposal()
		if err != nil {
			return nil, err
		}

		proposals[index] = proposalParsed
	}

	return proposals, nil
}

func (p ProposalWithID) ToProposal() (types.Proposal, error) {
	expiresAt, err := strconv.ParseInt(p.Proposal.Expiration.AtTime, 10, 64)
	if err != nil {
		return types.Proposal{}, err
	}

	return types.Proposal{
		ID:          strconv.Itoa(p.ID),
		Title:       p.Proposal.Title,
		Description: p.Proposal.Description,
		EndTime:     time.Unix(0, expiresAt),
		Status:      ParseProposalStatus(p.Proposal.Status),
	}, nil
}

func (p ProposalsResponse) ToTally() ([]types.TallyInfo, error) {
	tallyInfos := make([]types.TallyInfo, 0)

	for _, proposal := range p.Data.Proposals {
		proposalParsed, err := proposal.ToProposal()
		if err != nil {
			return []types.TallyInfo{}, err
		}

		if !proposalParsed.IsInVoting() {
			continue
		}

		yesVotes, err := math.LegacyNewDecFromStr(proposal.Proposal.Votes.Yes)
		if err != nil {
			return []types.TallyInfo{}, err
		}

		noVotes, err := math.LegacyNewDecFromStr(proposal.Proposal.Votes.No)
		if err != nil {
			return []types.TallyInfo{}, err
		}

		abstainVotes, err := math.LegacyNewDecFromStr(proposal.Proposal.Votes.Abstain)
		if err != nil {
			return []types.TallyInfo{}, err
		}

		totalVotes, err := math.LegacyNewDecFromStr(proposal.Proposal.TotalPower)
		if err != nil {
			return []types.TallyInfo{}, err
		}

		tallyInfos = append(tallyInfos, types.TallyInfo{
			Proposal: proposalParsed,
			Tally: types.Tally{
				{Option: "Yes", Voted: yesVotes},
				{Option: "No", Voted: noVotes},
				{Option: "Abstain", Voted: abstainVotes},
			},
			TotalVotingPower: totalVotes,
		})
	}

	return tallyInfos, nil
}
