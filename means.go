package main

import (
	"strings"
)

type meansCommand struct {
	max    int
	topics string
}

func (m *meansCommand) suggest(query string) ([]WordsResponse, error) {
	if len(query) < 3 {
		return []WordsResponse{}, nil
	}

	api := API()
	if err := api.Constrain(MeansLike, None); err != nil {
		return nil, err
	}
	if err := api.Limit(m.max); err != nil {
		return nil, err
	}
	if err := api.RelateToTopics(parseTopics(m.topics)); err != nil {
		return nil, err
	}

	return api.Complete(strings.ToLower(query))
}

func parseTopics(topicsStr string) []string {
	if topicsStr == "" {
		return []string{}
	}
	topics := strings.Split(topicsStr, ",")
	var result []string
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			result = append(result, topic)
		}
	}
	return result
} 