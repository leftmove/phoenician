package main

import (
	"strings"
)

type spellCommand struct {
	max    int
	topics string
}

func (s *spellCommand) suggest(query string) ([]WordsResponse, error) {
	if len(query) < 3 {
		return []WordsResponse{}, nil
	}

	api := API()
	if err := api.Constrain(SpellsLike, None); err != nil {
		return nil, err
	}
	if err := api.Limit(s.max); err != nil {
		return nil, err
	}
	if err := api.RelateToTopics(parseTopics(s.topics)); err != nil {
		return nil, err
	}

	return api.Complete(strings.ToLower(query))
} 