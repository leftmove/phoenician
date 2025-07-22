package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4FF")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1).
			Align(lipgloss.Center)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Align(lipgloss.Center).
				MarginBottom(2)

	resultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B8B8B"))

	numberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B8B8B"))

	noResultsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Italic(true)
)

type wordItem struct {
	word  string
	score int
}

func (i wordItem) FilterValue() string { return i.word }
func (i wordItem) Title() string       { return i.word }
func (i wordItem) Description() string { return fmt.Sprintf("Score: %d", i.score) }

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Tab    key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.Enter, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "search"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "autocomplete"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

type suggestionResponse struct {
	words []WordsResponse
}

type model struct {
	title           string
	description     string
	textInput       textinput.Model
	suggestionsList list.Model
	help            help.Model
	suggester       wordSuggester
	finalResults    []WordsResponse
	showResults     bool
	showSuggestions bool
	quitting        bool
	keys            keyMap
}

type wordSuggester interface {
	suggest(query string) ([]WordsResponse, error)
}

func initialModel(title, description string, suggester wordSuggester) model {
	ti := textinput.New()
	ti.Placeholder = "Start typing..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.Prompt = "> "

	items := []list.Item{}
	suggestionsList := list.New(items, list.NewDefaultDelegate(), 60, 20)
	suggestionsList.Title = ""
	suggestionsList.SetShowStatusBar(false)
	suggestionsList.SetFilteringEnabled(false)
	suggestionsList.SetShowHelp(false)
	suggestionsList.SetShowTitle(false)
	suggestionsList.InfiniteScrolling = true

	return model{
		title:           title,
		description:     description,
		textInput:       ti,
		suggestionsList: suggestionsList,
		help:            help.New(),
		suggester:       suggester,
		showResults:     false,
		showSuggestions: false,
		keys:            keys,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Enter):
			if m.showSuggestions && len(m.suggestionsList.Items()) > 0 {
				if selected, ok := m.suggestionsList.SelectedItem().(wordItem); ok {
					m.textInput.SetValue(selected.word)
				}
			}
			
			query := strings.TrimSpace(m.textInput.Value())
			if query != "" {
				results, err := m.suggester.suggest(query)
				if err == nil {
					m.finalResults = results
					m.showResults = true
					m.showSuggestions = false
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			if m.showSuggestions && len(m.suggestionsList.Items()) > 0 {
				if selected, ok := m.suggestionsList.SelectedItem().(wordItem); ok {
					m.textInput.SetValue(selected.word)
					return m, m.getSuggestions(selected.word)
				}
			}

		case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
			if m.showSuggestions {
				var cmd tea.Cmd
				m.suggestionsList, cmd = m.suggestionsList.Update(msg)
				return m, cmd
			}
		}

		// Always update text input
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		query := strings.TrimSpace(m.textInput.Value())
		if len(query) >= 3 {
			cmds = append(cmds, m.getSuggestions(query))
		} else {
			m.showSuggestions = false
			m.suggestionsList.SetItems([]list.Item{})
		}

	case suggestionResponse:
		items := make([]list.Item, len(msg.words))
		for i, word := range msg.words {
			items[i] = wordItem{word: word.Word, score: word.Score}
		}
		m.suggestionsList.SetItems(items)
		m.showSuggestions = len(items) > 0

	case tea.WindowSizeMsg:
		m.suggestionsList.SetWidth(msg.Width - 4)
	}

	return m, tea.Batch(cmds...)
}

func (m model) getSuggestions(query string) tea.Cmd {
	return func() tea.Msg {
		words, err := m.suggester.suggest(query)
		if err != nil {
			return suggestionResponse{words: []WordsResponse{}}
		}
		return suggestionResponse{words: words}
	}
}

func (m model) View() string {
	if m.quitting {
		return "\nGoodbye!\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(m.title))
	s.WriteString("\n")
	s.WriteString(descriptionStyle.Render(m.description))
	s.WriteString("\n")

	if m.showResults {
		s.WriteString(m.renderResults())
		return s.String()
	}

	s.WriteString(m.textInput.View())
	s.WriteString("\n")

	if len(m.textInput.Value()) > 0 && len(m.textInput.Value()) < 3 {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8B8B8B")).Render("Type at least 3 characters for suggestions..."))
		s.WriteString("\n")
	} else if m.showSuggestions && len(m.suggestionsList.Items()) > 0 {
		s.WriteString("\n")
		s.WriteString(m.suggestionsList.View())
	}

	s.WriteString("\n")
	helpView := m.help.View(m.keys)
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8B8B8B")).Render(helpView))

	return s.String()
}

func (m model) renderResults() string {
	var s strings.Builder
	
	s.WriteString("\n")
	if len(m.finalResults) == 0 {
		s.WriteString(noResultsStyle.Render("No results found."))
	} else {
		for i, word := range m.finalResults {
			number := numberStyle.Render(fmt.Sprintf("%d.", i+1))
			wordText := resultStyle.Render(word.Word)
			score := scoreStyle.Render(fmt.Sprintf("(%d)", word.Score))
			
			s.WriteString(fmt.Sprintf("  %s %s %s\n", number, wordText, score))
		}
	}
	s.WriteString("\n")
	
	return s.String()
}

func runApp(title, description string, suggester wordSuggester) error {
	p := tea.NewProgram(initialModel(title, description, suggester))
	_, err := p.Run()
	return err
}

var rootCmd = &cobra.Command{
	Use:   "phoenician",
	Short: "Find similar words",
	Long:  "Phoenician is a CLI tool for finding similar words using various search constraints.",
}

func init() {
	rootCmd.AddCommand(newMeansCommand())
	rootCmd.AddCommand(newSoundsCommand())
	rootCmd.AddCommand(newSpellCommand())
	rootCmd.AddCommand(newRelatesCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newMeansCommand() *cobra.Command {
	var maxResults int
	var topics string

	cmd := &cobra.Command{
		Use:   "means",
		Short: "Find words that mean like the input",
		Long:  "Start typing to get words that mean like the input",
		RunE: func(cmd *cobra.Command, args []string) error {
			suggester := &meansCommand{max: maxResults, topics: topics}
			return runApp("Phoenician", "Start typing to get words that mean like the input", suggester)
		},
	}

	cmd.Flags().IntVarP(&maxResults, "max", "m", 10, "Maximum number of results to show")
	cmd.Flags().StringVarP(&topics, "topics", "t", "", "Comma-separated list of topics to filter by. Ex: 'religion,mythology'")

	return cmd
}

func newSoundsCommand() *cobra.Command {
	var maxResults int
	var topics string

	cmd := &cobra.Command{
		Use:   "sounds",
		Short: "Find words that sound like the input",
		Long:  "Start typing to get words that sound like the input",
		RunE: func(cmd *cobra.Command, args []string) error {
			suggester := &soundsCommand{max: maxResults, topics: topics}
			return runApp("Phoenician", "Start typing to get words that sound like the input", suggester)
		},
	}

	cmd.Flags().IntVarP(&maxResults, "max", "m", 10, "Maximum number of results to show")
	cmd.Flags().StringVarP(&topics, "topics", "t", "", "Comma-separated list of topics to filter by. Ex: 'religion,mythology'")

	return cmd
}

func newSpellCommand() *cobra.Command {
	var maxResults int
	var topics string

	cmd := &cobra.Command{
		Use:   "spell",
		Short: "Find words that are spelled like the input",
		Long:  "Start typing to get words that are spelled like the input",
		RunE: func(cmd *cobra.Command, args []string) error {
			suggester := &spellCommand{max: maxResults, topics: topics}
			return runApp("Phoenician", "Start typing to get words that are spelled like the input", suggester)
		},
	}

	cmd.Flags().IntVarP(&maxResults, "max", "m", 10, "Maximum number of results to show")
	cmd.Flags().StringVarP(&topics, "topics", "t", "", "Comma-separated list of topics to filter by. Ex: 'religion,mythology'")

	return cmd
}

func newRelatesCommand() *cobra.Command {
	var maxResults int
	var topics string

	cmd := &cobra.Command{
		Use:   "relates [relation]",
		Short: "Find words related by a specific relation type",
		Long: `Find words related by a specific relation type.

Valid relation codes include the following.

  jja - Popular nouns modified by adjective (gradual → increase)
  jjb - Popular adjectives for noun (beach → sandy)
  syn - Synonyms (ocean → sea)
  trg - Trigger words (cow → milking)
  ant - Antonyms (late → early)
  spc - Kind of / hypernyms (gondola → boat)
  com - Comprises / holonyms (car → accelerator)
  par - Part of / meronyms (trunk → tree)
  bga - Frequent followers (wreak → havoc)
  bgb - Frequent predecessors (havoc → wreak)
  hom - Homophones (course → coarse)
  cns - Consonant match (sample → simple)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			relationCode := args[0]
			suggester := &relatesCommand{relation: relationCode, max: maxResults, topics: topics}
			description := fmt.Sprintf("Start typing to get words that are related by '%s'", getRelationDescription(relationCode))
			return runApp("Phoenician", description, suggester)
		},
	}

	cmd.Flags().IntVarP(&maxResults, "max", "m", 10, "Maximum number of results to show")
	cmd.Flags().StringVarP(&topics, "topics", "t", "", "Comma-separated list of topics to filter by. Ex: 'religion,mythology'")

	return cmd
}

