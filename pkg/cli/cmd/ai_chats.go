package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	aiChatsListPage       pageFlags
	aiChatsMessagesPage   pageFlags
	aiChatsCreateTitle    string
	aiChatsCreateTemplate string
	aiChatsContent        string
	aiChatsTemplate       string
)

// aiChatTemplateUsage names the two chat templates. The API defaults to COPILOT.
var aiChatTemplateUsage = fmt.Sprintf("chat template: %v", v3.AllowedENUMPROPERTIESTEMPLATEEnumValues)

var aiChatsCmd = &cobra.Command{
	Use:     "ai-chats",
	Aliases: []string{"ai-chat"},
	Short:   "Drive the copilot chats (Beta)",
	Long: `AI Chats - Talk to the Security Copilot (Beta)

Create a chat, send it a question and read the thread back. The copilot
answers from the security data of the organization: findings, assets and scan
results.

BETA: copilot chats are announced but the API contract is new and still
moving, so this group can still change.

COMMON WORKFLOWS:
  • Open a chat, ask, then read the thread:
    $ escape-cli ai-chats create --title "Investigate the login flow"
    $ escape-cli ai-chats message <chat-id> --content "Which critical issues are still open?"
    $ escape-cli ai-chats get <chat-id>

  • Walk an older thread:
    $ escape-cli ai-chats list
    $ escape-cli ai-chats messages <chat-id>`,
}

var aiChatsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the copilot chats of the caller (Beta)",
	Long: `List AI Chats - Most Recent First

List the copilot chats the API key belongs to. With --size or --cursor the
command returns one page as {items, nextCursor, totalCount}; otherwise it
walks every page.`,
	Example: `  escape-cli ai-chats list
  escape-cli ai-chats list --size 20 -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.AIChatSummary]{}) {
			return nil
		}

		if err := runPagedList(cmd, aiChatsListPage, func(ctx context.Context, cursor string, size int) ([]v3.AIChatSummary, *string, int, error) {
			return escape.ListAIChats(ctx, cursor, size)
		}, aiChatSummariesTable); err != nil {
			return fmt.Errorf("unable to list ai chats: %w", err)
		}

		return nil
	},
}

var aiChatsGetCmd = &cobra.Command{
	Use:     "get chat-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a chat and its most recent messages (Beta)",
	Long: `Get AI Chat - Read a Thread Without Paging

Fetch a copilot chat by ID. The response carries the first 100 messages of
the thread; use "ai-chats messages" to page through a longer history.`,
	Example: `  escape-cli ai-chats get 00000000-0000-0000-0000-000000000000
  escape-cli ai-chats get <chat-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetAIChat200Response{}) {
			return nil
		}

		chat, err := escape.GetAIChat(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get ai chat: %w", err)
		}

		out.Table(chat, func() []string {
			return aiChatMessagesTable(chat.GetMessages())
		})

		return nil
	},
}

var aiChatsMessagesCmd = &cobra.Command{
	Use:     "messages chat-id",
	Aliases: []string{"history"},
	Short:   "List the messages of a chat, oldest first (Beta)",
	Long: `List AI Chat Messages - Walk the Full Thread

List the messages of a copilot chat, oldest first. With --size or --cursor
the command returns one page as {items, nextCursor, totalCount}; otherwise it
walks every page.`,
	Example: `  escape-cli ai-chats messages <chat-id>
  escape-cli ai-chats messages <chat-id> --size 10 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(Page[v3.AIChatMessage]{}) {
			return nil
		}

		if err := runPagedList(cmd, aiChatsMessagesPage, func(ctx context.Context, cursor string, size int) ([]v3.AIChatMessage, *string, int, error) {
			return escape.ListAIChatMessages(ctx, args[0], cursor, size)
		}, aiChatMessagesTable); err != nil {
			return fmt.Errorf("unable to list ai chat messages: %w", err)
		}

		return nil
	},
}

var aiChatsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Open a new copilot chat (Beta)",
	Long: `Create AI Chat - Open a New Thread

Open a copilot chat with --title, optionally on one of the two templates. The
chat id it returns is what "ai-chats message", "ai-chats get" and
"ai-chats messages" take.`,
	Example: `  escape-cli ai-chats create --title "Investigate the login flow"
  escape-cli ai-chats create --title "Triage the backlog" --template DEFAULT`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.CreateAIChat200Response{}) {
			return nil
		}

		if aiChatsCreateTitle == "" {
			return errors.New("--title is required")
		}

		chat, err := escape.CreateAIChat(cmd.Context(), v3.CreateAIChatRequest{
			Title:    aiChatsCreateTitle,
			Template: aiChatTemplate(aiChatsCreateTemplate),
		})
		if err != nil {
			return fmt.Errorf("unable to create ai chat: %w", err)
		}

		out.Print(chat, fmt.Sprintf("Chat %s created: %s", chat.GetId(), chat.GetTitle()))

		return nil
	},
}

var aiChatsMessageCmd = &cobra.Command{
	Use:     "message chat-id",
	Aliases: []string{"send", "ask"},
	Short:   "Send a message to a chat (Beta)",
	Long: `Send AI Chat Message - Ask the Copilot

Post a user message to a copilot chat with --content. The copilot answer
arrives asynchronously, so what this returns is the user message the API
created, not the answer. Read the thread back with "ai-chats get" or
"ai-chats messages".`,
	Example: `  escape-cli ai-chats message <chat-id> --content "Which critical issues are still open?"
  escape-cli ai-chats message <chat-id> --content "Summarize the attack surface" --template DEFAULT`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateAIChatMessage200Response{}) {
			return nil
		}

		if aiChatsContent == "" {
			return errors.New("--content is required")
		}

		message, err := escape.CreateAIChatMessage(cmd.Context(), args[0], v3.CreateAIChatMessageRequest{
			Content:  aiChatsContent,
			Template: aiChatTemplate(aiChatsTemplate),
		})
		if err != nil {
			return fmt.Errorf("unable to send ai chat message: %w", err)
		}

		out.Table(message, func() []string {
			return []string{
				"MESSAGE ID\tROLE\tCONTENT\tCREATED AT",
				fmt.Sprintf("%s\t%s\t%s\t%s",
					message.GetId(),
					message.GetRole(),
					tableCell(message.GetContent()),
					aiChatDate(message.GetCreatedAt()),
				),
			}
		})

		return nil
	},
}

// aiChatTemplate maps the --template flag onto the generated enum. An empty
// flag keeps the field unset so the API applies its own default.
func aiChatTemplate(template string) *v3.ENUMPROPERTIESTEMPLATE {
	if template == "" {
		return nil
	}

	value := v3.ENUMPROPERTIESTEMPLATE(template)

	return &value
}

// aiChatDate renders a chat timestamp the way every other table does. The
// generated model decodes createdAt into a time.Time, and its String() form is
// not the RFC 3339 layout out.GetShortDate parses, so it has to be rendered
// first.
func aiChatDate(createdAt time.Time) string {
	return out.GetShortDate(createdAt.Format(time.RFC3339))
}

// tableCellReplacer keeps free text (chat titles, message bodies) on one row of
// the tab-separated pretty table. JSON and YAML output keep the raw values.
var tableCellReplacer = strings.NewReplacer("\t", " ", "\r\n", " ", "\n", " ", "\r", " ")

func tableCell(value string) string {
	return tableCellReplacer.Replace(value)
}

func aiChatSummariesTable(chats []v3.AIChatSummary) []string {
	rows := []string{"ID\tTITLE\tCREATED AT"}
	for _, chat := range chats {
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s",
			chat.GetId(),
			tableCell(chat.GetTitle()),
			aiChatDate(chat.GetCreatedAt()),
		))
	}

	return rows
}

func aiChatMessagesTable(messages []v3.AIChatMessage) []string {
	rows := []string{"ROLE\tCONTENT\tCREATED AT"}
	for _, message := range messages {
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s",
			message.GetRole(),
			tableCell(message.GetContent()),
			aiChatDate(message.GetCreatedAt()),
		))
	}

	return rows
}

func init() {
	aiChatsListPage.bind(aiChatsListCmd)
	aiChatsMessagesPage.bind(aiChatsMessagesCmd)

	aiChatsCreateCmd.Flags().StringVar(&aiChatsCreateTitle, "title", "", "title of the chat")
	aiChatsCreateCmd.Flags().StringVar(&aiChatsCreateTemplate, "template", "", aiChatTemplateUsage)
	markMCPRequired(aiChatsCreateCmd.Flags(), "title")
	aiChatsMessageCmd.Flags().StringVar(&aiChatsContent, "content", "", "content of the user message")
	aiChatsMessageCmd.Flags().StringVar(&aiChatsTemplate, "template", "", aiChatTemplateUsage)
	markMCPRequired(aiChatsMessageCmd.Flags(), "content")

	aiChatsCmd.AddCommand(aiChatsListCmd, aiChatsGetCmd, aiChatsMessagesCmd, aiChatsCreateCmd, aiChatsMessageCmd)
	rootCmd.AddCommand(aiChatsCmd)
}
