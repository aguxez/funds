package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputState int

const (
	PENDING_INPUT InputState = iota
	INPUT_SENT
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder())

// category represents a group of elements with a global percentage, for example, fixed costs
type category struct {
	name       string
	percentage int
}

type model struct {
	input           textinput.Model
	categories      []category
	inputState      InputState
	categoriesTable table.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Euro amount"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 20

	categories := []category{
		{name: "Fixed costs", percentage: 50},
		{name: "Investments", percentage: 25},
		{name: "Savings", percentage: 10},
		{name: "Guilt-free spending", percentage: 15},
	}

	return model{
		categories: categories,
		input:      ti,
		inputState: PENDING_INPUT,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.inputState = INPUT_SENT

			// Build the tables we'll use in the View
			// Categories first
			categoryColumns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "Percentage", Width: 10},
				{Title: "Amount", Width: 10},
			}

			intInput, err := strconv.Atoi(m.input.Value())
			if err != nil {
				return m, tea.Quit
			}

			var categoryRows []table.Row
			for _, category := range m.categories {
				categoryRows = append(
					categoryRows,
					table.Row{
						category.name,
						strconv.Itoa(category.percentage),
						// At this point we have the complete value
						strconv.Itoa(intInput * category.percentage / 100),
					},
				)
			}

			categoriesTable := table.New(
				table.WithColumns(categoryColumns),
				table.WithRows(categoryRows),
				table.WithHeight(len(m.categories)),
			)

			style := table.DefaultStyles()
			style.Header = style.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("182")).
				BorderBottom(true).
				Bold(false)

			// Overrides the selected effect in the table to an empty one.
			style.Selected = style.Selected.Foreground(lipgloss.Color(""))

			categoriesTable.SetStyles(style)

			m.categoriesTable = categoriesTable

			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m model) View() string {
	var output string

	if m.inputState == PENDING_INPUT {
		// Header
		output += "What is the received income?\n\n"
		output += m.input.View()
		// Footer
		output += "\n\nPress 'esc' to quit\n"
	}

	if m.inputState == INPUT_SENT {
		_, err := strconv.Atoi(m.input.Value())

		if err != nil {
			return "There was an error reading your input"
		}

		// We have to calculate the spendings/amounts based on percentages and fixed costs.
		// Show categories and their percentages or amounts.
		// These tables were built when initializing the model.
		output = baseStyle.Render(m.categoriesTable.View()) + "\n"
	}

	return output
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("There has been an error: %v", err)
		os.Exit(1)
	}
}
