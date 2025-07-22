package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrTooManyResults   = errors.New("number entered is too high")
	ErrRelationRequired = errors.New("the extra word parameter is required if the relatesLike search constraint is chosen")
	ErrWordRequired     = errors.New("a three-letter identifier code is required to use relatesLike; this code refers to the type of relation")
	ErrBadURL          = errors.New("error occurred building URL")
	ErrTooManyTopics   = errors.New("too many or too few topics specified, max is 1000")
)

type SearchConstraint int

const (
	MeansLike SearchConstraint = iota
	SoundsLike
	SpellsLike
	RelatesLike
)

type RelationConstraint int

const (
	NounByAdjective RelationConstraint = iota
	AdjectiveByNoun
	Synonyms
	Trigger
	Antonyms
	KindOf
	Comprises
	PartOf
	FrequentFollowers
	FrequentPredecessors
	Homophones
	Consonant
	None
)

func (r RelationConstraint) Code() string {
	switch r {
	case NounByAdjective:
		return "jja"
	case AdjectiveByNoun:
		return "jjb"
	case Synonyms:
		return "syn"
	case Trigger:
		return "trg"
	case Antonyms:
		return "ant"
	case KindOf:
		return "spc"
	case Comprises:
		return "com"
	case PartOf:
		return "par"
	case FrequentFollowers:
		return "bga"
	case FrequentPredecessors:
		return "bgb"
	case Homophones:
		return "hom"
	case Consonant:
		return "cns"
	default:
		return "none"
	}
}

func RelationFromCode(code string) RelationConstraint {
	switch code {
	case "jja":
		return NounByAdjective
	case "jjb":
		return AdjectiveByNoun
	case "syn":
		return Synonyms
	case "trg":
		return Trigger
	case "ant":
		return Antonyms
	case "spc":
		return KindOf
	case "com":
		return Comprises
	case "par":
		return PartOf
	case "bga":
		return FrequentFollowers
	case "bgb":
		return FrequentPredecessors
	case "hom":
		return Homophones
	case "cns":
		return Consonant
	default:
		return None
	}
}

type WordsResponse struct {
	Word  string `json:"word"`
	Score int    `json:"score"`
}

type APIProperties struct {
	baseURL    string
	constraint string
	query      string
	maxResults int
	topics     []string
}

func API() *APIProperties {
	return &APIProperties{
		baseURL:    "https://api.datamuse.com/words",
		constraint: "sp",
		maxResults: 12,
		topics:     []string{},
	}
}

func (api *APIProperties) Query(query string) *APIProperties {
	api.query = query
	return api
}

func (api *APIProperties) Constrain(constraint SearchConstraint, relation RelationConstraint) error {
	switch constraint {
	case MeansLike:
		api.constraint = "ml"
	case SoundsLike:
		api.constraint = "sl"
	case SpellsLike:
		api.constraint = "sp"
	case RelatesLike:
		if relation == None {
			return ErrRelationRequired
		}
		api.constraint = "rel_" + relation.Code()
	}
	return nil
}

func (api *APIProperties) Limit(max int) error {
	if max < 1 || max > 1000 {
		return ErrTooManyResults
	}
	api.maxResults = max
	return nil
}

func (api *APIProperties) RelateToTopics(topics []string) error {
	if len(topics) > 1000 {
		return ErrTooManyTopics
	}
	api.topics = topics
	return nil
}

func (api *APIProperties) Complete(query string) ([]WordsResponse, error) {
	baseURL, err := url.Parse(api.baseURL)
	if err != nil {
		return nil, ErrBadURL
	}

	params := url.Values{}
	params.Set(api.constraint, query)
	params.Set("max", strconv.Itoa(api.maxResults))
	if len(api.topics) > 0 {
		params.Set("topics", strings.Join(api.topics, ","))
	}

	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var words []WordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&words); err != nil {
		return nil, err
	}

	return words, nil
} 