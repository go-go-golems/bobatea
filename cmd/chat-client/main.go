package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-go-golems/bobatea/pkg/chat"
	conversationui "github.com/go-go-golems/bobatea/pkg/chat/conversation"
	"github.com/go-go-golems/bobatea/pkg/conversation"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
	parentID   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chat-client",
		Short: "HTTP client for Bobatea chat backend and user interface",
		Long:  `A command-line tool to interact with the Bobatea chat HTTP backend and user interface.`,
	}

	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	rootCmd.PersistentFlags().StringVarP(&parentID, "parent", "p", "", "Parent message ID")

	backendCmd := &cobra.Command{
		Use:   "backend",
		Short: "Interact with the chat backend",
	}
	backendCmd.AddCommand(newStartCmd())
	backendCmd.AddCommand(newCompletionCmd())
	backendCmd.AddCommand(newStatusCmd())
	backendCmd.AddCommand(newDoneCmd())
	backendCmd.AddCommand(newErrorCmd())
	backendCmd.AddCommand(newFinishCmd())
	backendCmd.AddCommand(newGetStatusCmd())

	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Interact with the user interface",
	}
	userCmd.AddCommand(newToggleHelpCmd())
	userCmd.AddCommand(newUnfocusMessageCmd())
	userCmd.AddCommand(newQuitCmd())
	userCmd.AddCommand(newFocusMessageCmd())
	userCmd.AddCommand(newSelectNextMessageCmd())
	userCmd.AddCommand(newSelectPrevMessageCmd())
	userCmd.AddCommand(newSubmitMessageCmd())
	userCmd.AddCommand(newCopyToClipboardCmd())
	userCmd.AddCommand(newCopyLastResponseToClipboardCmd())
	userCmd.AddCommand(newCopyLastSourceBlocksToClipboardCmd())
	userCmd.AddCommand(newCopySourceBlocksToClipboardCmd())
	userCmd.AddCommand(newSaveToFileCmd())
	userCmd.AddCommand(newCancelCompletionCmd())
	userCmd.AddCommand(newDismissErrorCmd())
	userCmd.AddCommand(newInputTextCmd())

	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(userCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Backend-related command functions
func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start a new chat session",
		Run: func(cmd *cobra.Command, args []string) {
			msg := conversationui.StreamStartMsg{
				StreamMetadata: conversationui.StreamMetadata{
					ID:       conversation.NewNodeID(),
					ParentID: stringToNodeID(parentID),
				},
			}
			sendRequest("start", msg)
		},
	}
}

func newCompletionCmd() *cobra.Command {
	var completion string
	cmd := &cobra.Command{
		Use:   "completion [delta]",
		Short: "Send a completion message",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			delta := strings.Join(args, " ")
			msg := conversationui.StreamCompletionMsg{
				StreamMetadata: conversationui.StreamMetadata{
					ID:       conversation.NewNodeID(),
					ParentID: stringToNodeID(parentID),
				},
				Delta: delta,
			}

			if completion != "" {
				msg.Completion = completion + delta
			}

			sendRequest("completion", msg)
		},
	}
	cmd.Flags().StringVarP(&completion, "completion", "c", "", "Completion text")
	return cmd
}

func newStatusCmd() *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Send a status update",
		Run: func(cmd *cobra.Command, args []string) {
			msg := conversationui.StreamStatusMsg{
				StreamMetadata: conversationui.StreamMetadata{
					ID:       conversation.NewNodeID(),
					ParentID: stringToNodeID(parentID),
				},
				Text: status,
			}
			sendRequest("status-update", msg)
		},
	}
	cmd.Flags().StringVarP(&status, "status", "t", "", "Status text")
	return cmd
}

func newDoneCmd() *cobra.Command {
	var completion string
	cmd := &cobra.Command{
		Use:   "done",
		Short: "Send a done message",
		Run: func(cmd *cobra.Command, args []string) {
			msg := conversationui.StreamDoneMsg{
				StreamMetadata: conversationui.StreamMetadata{
					ID:       conversation.NewNodeID(),
					ParentID: stringToNodeID(parentID),
				},
				Completion: completion,
			}
			sendRequest("done", msg)
		},
	}
	cmd.Flags().StringVarP(&completion, "completion", "c", "", "Final completion text")
	return cmd
}

func newErrorCmd() *cobra.Command {
	var errMsg string
	cmd := &cobra.Command{
		Use:   "error",
		Short: "Send an error message",
		Run: func(cmd *cobra.Command, args []string) {
			msg := conversationui.StreamCompletionError{
				StreamMetadata: conversationui.StreamMetadata{
					ID:       conversation.NewNodeID(),
					ParentID: stringToNodeID(parentID),
				},
				Err: fmt.Errorf("%s", errMsg),
			}
			sendRequest("error", msg)
		},
	}
	cmd.Flags().StringVarP(&errMsg, "message", "m", "", "Error message")
	return cmd
}

func newFinishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "finish",
		Short: "Send a finish message",
		Run: func(cmd *cobra.Command, args []string) {
			msg := chat.BackendFinishedMsg{}
			sendRequest("finish", msg)
		},
	}
}

func newGetStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-status",
		Short: "Get the current status of the chat backend",
		Run: func(cmd *cobra.Command, args []string) {
			getStatus()
		},
	}
}

// User interface command functions
func newToggleHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "toggle-help",
		Short: "Toggle help display",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("toggle-help", nil)
		},
	}
}

func newUnfocusMessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unfocus-message",
		Short: "Unfocus the current message",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("unfocus-message", nil)
		},
	}
}

func newQuitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quit",
		Short: "Quit the application",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("quit", nil)
		},
	}
}

func newFocusMessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "focus-message",
		Short: "Focus on a message",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("focus-message", nil)
		},
	}
}

func newSelectNextMessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "select-next-message",
		Short: "Select the next message",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("select-next-message", nil)
		},
	}
}

func newSelectPrevMessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "select-prev-message",
		Short: "Select the previous message",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("select-prev-message", nil)
		},
	}
}

func newSubmitMessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "submit-message",
		Short: "Submit a new message",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("submit-message", nil)
		},
	}
}

func newCopyToClipboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-to-clipboard",
		Short: "Copy selected content to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("copy-to-clipboard", nil)
		},
	}
}

func newCopyLastResponseToClipboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-last-response",
		Short: "Copy the last response to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("copy-last-response-to-clipboard", nil)
		},
	}
}

func newCopyLastSourceBlocksToClipboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-last-source-blocks",
		Short: "Copy the last source blocks to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("copy-last-source-blocks-to-clipboard", nil)
		},
	}
}

func newCopySourceBlocksToClipboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-source-blocks",
		Short: "Copy all source blocks to clipboard",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("copy-source-blocks-to-clipboard", nil)
		},
	}
}

func newSaveToFileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save-to-file",
		Short: "Save the conversation to a file",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("save-to-file", nil)
		},
	}
}

func newCancelCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-completion",
		Short: "Cancel the current completion",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("cancel-completion", nil)
		},
	}
}

func newDismissErrorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dismiss-error",
		Short: "Dismiss the current error",
		Run: func(cmd *cobra.Command, args []string) {
			sendUserRequest("dismiss-error", nil)
		},
	}
}

func newInputTextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "input-text <text>",
		Short: "Input text to the chat",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			text := args[0]
			sendUserRequest("input-text", map[string]string{"text": text})
		},
	}
}

func sendRequest(endpoint string, msg interface{}) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/backend/%s", serverAddr, endpoint), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned status code %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Request sent successfully")
}

func sendUserRequest(endpoint string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/user/%s", serverAddr, endpoint), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned status code %d\n", resp.StatusCode)
		return
	}

	fmt.Println("Request sent successfully")
}

func getStatus() {
	resp, err := http.Get(fmt.Sprintf("%s/backend/status", serverAddr))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Server returned status code %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	var status struct {
		Status      string                              `json:"status"`
		Messages    []*conversation.Message             `json:"messages"`
		LastMessage *conversationui.StreamCompletionMsg `json:"last_message,omitempty"`
		LastError   string                              `json:"last_error,omitempty"`
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", status.Status)
	fmt.Printf("Number of messages: %d\n", len(status.Messages))
	if status.LastMessage != nil {
		fmt.Printf("Last message delta: %s\n", status.LastMessage.Delta)
		fmt.Printf("Last message completion: %s\n", status.LastMessage.Completion)
	}
	if status.LastError != "" {
		fmt.Printf("Last error: %s\n", status.LastError)
	}
}

// Helper function to convert string to NodeID
func stringToNodeID(s string) conversation.NodeID {
	if s == "" {
		return conversation.NullNode
	}
	id, err := uuid.Parse(s)
	if err != nil {
		fmt.Printf("Warning: Invalid UUID format for parent ID. Using NullNode instead.\n")
		return conversation.NullNode
	}
	return conversation.NodeID(id)
}
