package main

import (
	"github.com/rs/zerolog"
)

type StateGenerator struct {
	Logger zerolog.Logger
	Chains []Chain
}

func NewStateGenerator(logger *zerolog.Logger, chains []Chain) *StateGenerator {
	return &StateGenerator{
		Logger: logger.With().Str("component", "state_generator").Logger(),
		Chains: chains,
	}
}

func (g *StateGenerator) GetState(oldState State) State {
	state := NewState()

	for _, chain := range g.Chains {
		g.Logger.Info().Str("name", chain.Name).Msg("Processing a chain")

		rpc := NewRPC(chain.LCDEndpoints, g.Logger)

		proposals, err := rpc.GetAllProposals()
		if err != nil {
			g.Logger.Warn().Err(err).Msg("Error processing proposals")
			state.SetChainProposalsError(chain.Name, err)
			continue
		}

		g.Logger.Info().Int("len", len(proposals)).Msg("Got proposals")

		for _, proposal := range proposals {
			g.Logger.Trace().
				Str("name", chain.Name).
				Str("proposal", proposal.ProposalID).
				Msg("Processing a proposal")

			for _, wallet := range chain.Wallets {
				g.Logger.Trace().
					Str("name", chain.Name).
					Str("proposal", proposal.ProposalID).
					Str("wallet", wallet).
					Msg("Processing wallet vote")

				oldVote, _, found := oldState.GetVoteAndProposal(chain.Name, proposal.ProposalID, wallet)
				voteResponse, err := rpc.GetVote(proposal.ProposalID, wallet)

				if found && oldVote.HasVoted() && voteResponse.Vote == nil {
					g.Logger.Trace().
						Str("chain", chain.Name).
						Str("proposal", proposal.ProposalID).
						Str("wallet", wallet).
						Msg("Wallet has voted and there's no vote in the new state - using old vote")

					state.SetVote(
						chain.Name,
						proposal,
						wallet,
						oldVote,
					)

					continue
				}

				proposalVote := ProposalVote{}

				if err != nil {
					proposalVote.Error = err.Error()
				} else {
					proposalVote.Vote = voteResponse.Vote
				}

				state.SetVote(
					chain.Name,
					proposal,
					wallet,
					proposalVote,
				)
			}
		}
	}

	return state
}