package cmd

import (
	"context"
	"net/http"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

const aiChatID = "00000000-0000-0000-0000-0000000000c0"

func aiChatFixture(chatID string) map[string]any {
	return map[string]any{
		"id":        chatID,
		"title":     "Investigate the login flow",
		"createdAt": "2026-10-01T10:00:00Z",
	}
}

// TestAIChatsCommandsPrintDeclaredDocument walks every leaf of the ai-chats
// group and checks each prints its declared response.
func TestAIChatsCommandsPrintDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/ai-chats":
			writeJSON(t, w, map[string]any{"data": []map[string]any{aiChatFixture(aiChatID)}, "totalCount": 1})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/ai-chats/"+aiChatID:
			writeJSON(t, w, map[string]any{
				"id":                 aiChatID,
				"title":              "Investigate the login flow",
				"createdAt":          "2026-10-01T10:00:00Z",
				"messages":           []map[string]any{{"id": "m1", "role": "USER", "content": "Which issues are open?", "createdAt": "2026-10-01T10:01:00Z"}},
				"messagesTotalCount": 1,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/ai-chats/"+aiChatID+"/messages":
			writeJSON(t, w, map[string]any{"data": []map[string]any{
				{"id": "m1", "role": "USER", "content": "Which issues are open?", "createdAt": "2026-10-01T10:01:00Z"},
			}, "totalCount": 1})
		case r.Method == http.MethodPost && r.URL.Path == "/v3/ai-chats":
			writeJSON(t, w, aiChatFixture(aiChatID))
		case r.Method == http.MethodPost && r.URL.Path == "/v3/ai-chats/"+aiChatID+"/messages":
			writeJSON(t, w, map[string]any{"id": "m2", "role": "USER", "content": "And now?", "createdAt": "2026-10-01T10:05:00Z"})
		default:
			http.NotFound(w, r)
		}
	})

	aiChatsCreateTitle = "Investigate the login flow"
	aiChatsContent = "Which issues are open?"
	t.Cleanup(func() {
		aiChatsCreateTitle = ""
		aiChatsContent = ""
	})

	tests := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "list",
			cmd:  aiChatsListCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var chats []v3.AIChatSummary
				mustUnmarshalOne(t, stdout, &chats)
				if len(chats) != 1 || chats[0].GetId() != aiChatID {
					t.Fatalf("chats = %#v", chats)
				}
			},
		},
		{
			name: "get",
			cmd:  aiChatsGetCmd,
			args: []string{aiChatID},
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var chat v3.GetAIChat200Response
				mustUnmarshalOne(t, stdout, &chat)
				if chat.GetId() != aiChatID || len(chat.GetMessages()) != 1 {
					t.Fatalf("chat = %#v", chat)
				}

				if chat.GetMessagesTotalCount() != 1 {
					t.Fatalf("messagesTotalCount = %d", chat.GetMessagesTotalCount())
				}
			},
		},
		{
			name: "messages",
			cmd:  aiChatsMessagesCmd,
			args: []string{aiChatID},
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var messages []v3.AIChatMessage
				mustUnmarshalOne(t, stdout, &messages)
				if len(messages) != 1 || messages[0].GetRole() != "USER" {
					t.Fatalf("messages = %#v", messages)
				}
			},
		},
		{
			name: "create",
			cmd:  aiChatsCreateCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var chat v3.CreateAIChat200Response
				mustUnmarshalOne(t, stdout, &chat)
				if chat.GetId() != aiChatID || chat.GetTitle() != "Investigate the login flow" {
					t.Fatalf("chat = %#v", chat)
				}
			},
		},
		{
			name: "message",
			cmd:  aiChatsMessageCmd,
			args: []string{aiChatID},
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var message v3.CreateAIChatMessage200Response
				mustUnmarshalOne(t, stdout, &message)
				if message.GetId() != "m2" || message.GetContent() != "And now?" {
					t.Fatalf("message = %#v", message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(context.Background())
			tt.cmd.SetIn(nil)

			stdout, stderr, err := captureJSONCommand(t, func() error {
				return tt.cmd.RunE(tt.cmd, tt.args)
			})
			if err != nil {
				t.Fatalf("%s: %v\nstderr: %s", tt.name, err, stderr)
			}

			tt.check(t, stdout)
		})
	}
}

func TestAIChatsCreateRequiresATitle(t *testing.T) {
	aiChatsCreateTitle = ""
	t.Cleanup(func() { aiChatsCreateTitle = "" })

	err := aiChatsCreateCmd.RunE(aiChatsCreateCmd, nil)
	if err == nil || err.Error() != "--title is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestAIChatsMessageRequiresContent(t *testing.T) {
	aiChatsContent = ""
	t.Cleanup(func() { aiChatsContent = "" })

	err := aiChatsMessageCmd.RunE(aiChatsMessageCmd, []string{aiChatID})
	if err == nil || err.Error() != "--content is required" {
		t.Fatalf("err = %v", err)
	}
}
