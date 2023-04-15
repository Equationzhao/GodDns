package tui

import (
	"fmt"

	"GodDns/core"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type sModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
	process  []tea.Cmd
}

func initialModel(cmds ...tea.Cmd) sModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return sModel{spinner: s, process: cmds}
}

func (m sModel) Init() tea.Cmd {
	m.process = append(m.process, m.spinner.Tick)
	return tea.Batch(m.process...)
}

func (m sModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	case struct{}:
		m.quitting = true
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m sModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s executing\n\n", m.spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
}

func ShowSpinner(cmds ...tea.Cmd) {
	p := tea.NewProgram(initialModel(cmds...))
	if _, err := p.Run(); err != nil {
		core.Errchan <- err
	}
}
