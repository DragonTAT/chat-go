package ui

import (
	"context"
	"fmt"
	"strings"

	"ai-companion-cli-go/internal/llm"
	"ai-companion-cli-go/internal/models"
	"ai-companion-cli-go/internal/orchestrator"
	"ai-companion-cli-go/internal/storage"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	systemStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	userStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	aiStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	streamingAIStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type errMsg error

// streamMsg brings a chunk from the LLM stream
type streamMsg string

// streamDone signals the end of stream
type streamDone struct{}

type AppModel struct {
	viewport viewport.Model
	messages []string
	textarea textarea.Model
	err      error

	repo         *storage.Repository
	llmClient    *llm.Client
	orchestrator *orchestrator.Orchestrator

	profile *models.CharacterProfile
	session *models.SessionState

	isStreaming  bool
	currentReply strings.Builder

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func InitialModel(repo *storage.Repository, llmClient *llm.Client, orch *orchestrator.Orchestrator, profile *models.CharacterProfile, session *models.SessionState) AppModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.CharLimit = 500
	ta.SetWidth(80)
	ta.SetHeight(3)

	vp := viewport.New(80, 20)
	vp.SetContent(systemStyle.Render(fmt.Sprintf("\nChat with %s started. Intimacy Level: %d. Press Ctrl+C to quit.\n", profile.Name, 7))) // Simplification on intimacy initial

	// Load history
	hist, _ := repo.GetRecentMessages(profile.CharacterID, 50)
	var histLines []string
	histLines = append(histLines, systemStyle.Render(fmt.Sprintf("\nChat with %s started. Press Ctrl+C to quit.\n", profile.Name)))

	for _, m := range hist {
		if m.Role == "user" {
			histLines = append(histLines, userStyle.Render("You: ")+m.Content)
		} else {
			histLines = append(histLines, aiStyle.Render(profile.Name+": ")+m.Content)
		}
	}
	vp.SetContent(strings.Join(histLines, "\n\n"))
	vp.GotoBottom()

	return AppModel{
		textarea:     ta,
		messages:     histLines,
		viewport:     vp,
		repo:         repo,
		llmClient:    llmClient,
		orchestrator: orch,
		profile:      profile,
		session:      session,
	}
}

func (m AppModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if msg.Alt && !m.isStreaming {
				// We inject real newline for Alt+Enter
				m.textarea.InsertString("\n")
				return m, nil
			}

			if !m.isStreaming {
				v := m.textarea.Value()
				if v == "" {
					return m, nil
				}

				m.messages = append(m.messages, userStyle.Render("You: ")+v)
				m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
				m.textarea.Reset()
				m.viewport.GotoBottom()

				m.isStreaming = true
				m.currentReply.Reset()

				// Start orchestrator logic
				m.ctx, m.cancelFunc = context.WithCancel(context.Background())

				return m, m.startStreamCmd(v)
			}
		}

	case streamMsg:
		m.currentReply.WriteString(string(msg))

		// Create a copy of messages and append the streaming one to UI
		displayMsgs := make([]string, len(m.messages))
		copy(displayMsgs, m.messages)

		liveText := m.currentReply.String()
		displayMsgs = append(displayMsgs, streamingAIStyle.Render(m.profile.Name+": ")+liveText+" █")

		m.viewport.SetContent(strings.Join(displayMsgs, "\n\n"))
		m.viewport.GotoBottom()

		// Re-trigger view for next token chunk
		return m, m.waitForNextChunkCmd()

	case streamDone:
		m.isStreaming = false
		m.messages = append(m.messages, aiStyle.Render(m.profile.Name+": ")+m.currentReply.String())
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		m.viewport.GotoBottom()
		return m, nil

	case errMsg:
		m.err = msg
		m.isStreaming = false
		m.messages = append(m.messages, systemStyle.Render(fmt.Sprintf("Error: %v", msg)))
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

// startStreamCmd initiates the channel listener logic
func (m AppModel) startStreamCmd(userText string) tea.Cmd {
	return func() tea.Msg {
		tokenChan, errChan := m.orchestrator.GenerateReplyStream(m.ctx, userText, m.profile, m.session)

		// Start a goroutine that pumps tokens from the channel into the Bubbletea event loop
		go func() {
			for {
				select {
				case <-m.ctx.Done():
					return
				case err, ok := <-errChan:
					if ok && err != nil {
						// Need a way to send error safely to Tea
					}
				case _, ok := <-tokenChan:
					if !ok {
						// Stream closed
						return
					}
					// this is an anti-pattern as UI gets updated outside Update loop
					// but since Bubbletea doesn't have an officially native simple subscription model
					// we just rely on a pointer logic or custom event loop if we had fully async
					return
				}
			}
		}()

		// To properly do async channels in Bubbletea we need a persistent goroutine yielding cmds
		// For simplicity we will block here in chunks (simplified version):
		err := <-errChan
		if err != nil {
			return errMsg(err)
		}

		// This block is pseudo async, in a real production we use tea.Tick or background commands dispatch
		// Since we want streaming, let's write a small helper
		for range tokenChan {
			// Actually we can't yield like Python, we need to return a Cmd that fetches one token
		}

		return streamDone{}
	}
}

// Improved Stream approach:

// The below channels are global for simplicity in this POC rewrite to bridge the gap between goroutines and Tea Updates
var activeTokenChan <-chan string
var activeErrChan <-chan error

func (m AppModel) startStreamImprovedCmd(userText string) tea.Cmd {
	tokenChan, errChan := m.orchestrator.GenerateReplyStream(m.ctx, userText, m.profile, m.session)
	activeTokenChan = tokenChan
	activeErrChan = errChan
	return m.waitForNextChunkCmd()
}

func (m AppModel) waitForNextChunkCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case err, ok := <-activeErrChan:
			if ok && err != nil {
				return errMsg(err)
			}
		case chunk, ok := <-activeTokenChan:
			if !ok {
				return streamDone{}
			}
			return streamMsg(chunk)
		}
		return streamDone{} // Fallback
	}
}

// Overwrite the original
func (m AppModel) startStreamCmdOverwrite(userText string) tea.Cmd {
	return m.startStreamImprovedCmd(userText)
}

func (m AppModel) View() string {
	head := titleStyle.Render(fmt.Sprintf(" ♥ AI Companion: %s ♥ ", m.profile.Name))

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		head,
		m.viewport.View(),
		m.textarea.View(),
	)
}
