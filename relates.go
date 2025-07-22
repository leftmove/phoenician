package main

import (
	"fmt"
	"strings"
)

type relatesCommand struct {
	relation string
	max      int
	topics   string
}

func (r *relatesCommand) suggest(query string) ([]WordsResponse, error) {
	if len(query) < 3 {
		return []WordsResponse{}, nil
	}

	relation := RelationFromCode(r.relation)
	if relation == None {
		return nil, fmt.Errorf("invalid relation code: %s", r.relation)
	}

	api := API()
	if err := api.Constrain(RelatesLike, relation); err != nil {
		return nil, err
	}
	if err := api.Limit(r.max); err != nil {
		return nil, err
	}
	if err := api.RelateToTopics(parseTopics(r.topics)); err != nil {
		return nil, err
	}

	return api.Complete(strings.ToLower(query))
}

func getRelationDescription(relationCode string) string {
	relation := RelationFromCode(relationCode)
	switch relation {
	case NounByAdjective:
		return "Popular nouns modified by adjective (gradual → increase)"
	case AdjectiveByNoun:
		return "Popular adjectives for noun (beach → sandy)"
	case Synonyms:
		return "Synonyms (ocean → sea)"
	case Trigger:
		return "Trigger words (cow → milking)"
	case Antonyms:
		return "Antonyms (late → early)"
	case KindOf:
		return "Kind of / hypernyms (gondola → boat)"
	case Comprises:
		return "Comprises / holonyms (car → accelerator)"
	case PartOf:
		return "Part of / meronyms (trunk → tree)"
	case FrequentFollowers:
		return "Frequent followers (wreak → havoc)"
	case FrequentPredecessors:
		return "Frequent predecessors (havoc → wreak)"
	case Homophones:
		return "Homophones (course → coarse)"
	case Consonant:
		return "Consonant match (sample → simple)"
	default:
		return "Unknown relation"
	}
} 