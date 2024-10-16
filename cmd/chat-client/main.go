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
		Short: "HTTP client for Bobatea chat backend",
		Long:  `A command-line tool to interact with the Bobatea chat HTTP backend.`,
	}

	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	rootCmd.PersistentFlags().StringVarP(&parentID, "parent", "p", "", "Parent message ID")

	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newDoneCmd())
	rootCmd.AddCommand(newErrorCmd())
	rootCmd.AddCommand(newFinishCmd())
	rootCmd.AddCommand(newGetStatusCmd()) // Add the new get-status command

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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

func getStatus() {
	resp, err := http.Get(fmt.Sprintf("%s/status", serverAddr))
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

func sendRequest(endpoint string, msg interface{}) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/%s", serverAddr, endpoint), "application/json", bytes.NewBuffer(jsonData))
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

// New helper function to convert string to NodeID
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
