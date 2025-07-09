package chat

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tea "github.com/charmbracelet/bubbletea"

	geppetto_conversation "github.com/go-go-golems/geppetto/pkg/conversation"
)

// MockBackend for testing
type MockBackend struct {
	finished bool
}

func (m *MockBackend) IsFinished() bool {
	return m.finished
}

func (m *MockBackend) Start(ctx context.Context, msgs []*geppetto_conversation.Message) (tea.Cmd, error) {
	return func() tea.Msg {
		return nil
	}, nil
}

func (m *MockBackend) Kill() {}

func (m *MockBackend) Interrupt() {}

func TestFilePickerCallback(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "chat_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a mock conversation manager
	manager := geppetto_conversation.NewManager()
	
	// Create a mock backend
	backend := &MockBackend{finished: true}

	// Create a callback that writes "hello" to the file
	callback := func(path string) error {
		return os.WriteFile(path, []byte("hello"), 0644)
	}

	// Create model with callback
	model := InitialModel(manager, backend, WithFilePickerCallback(callback))

	// Test that callback is set
	assert.NotNil(t, model.filePickerCallback)

	// Create a test file path
	testFile := filepath.Join(tmpDir, "test.txt")

	// Simulate file selection
	resultModel, _ := model.handleFileSelection(testFile)

	// Verify file was created with "hello" content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))

	// Verify model returned successfully
	assert.NotNil(t, resultModel)
}

func TestFilePickerCallbackWithError(t *testing.T) {
	// Create a mock conversation manager
	manager := geppetto_conversation.NewManager()
	
	// Create a mock backend
	backend := &MockBackend{finished: true}

	// Create a callback that returns an error
	callback := func(path string) error {
		return os.ErrPermission
	}

	// Create model with callback
	model := InitialModel(manager, backend, WithFilePickerCallback(callback))

	// Simulate file selection
	resultModel, cmd := model.handleFileSelection("/some/path")

	// Verify error is returned
	assert.NotNil(t, cmd)
	
	// Execute the command to get the error message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	assert.True(t, ok)
	assert.Equal(t, os.ErrPermission, errorMsg)

	// Verify model returned successfully
	assert.NotNil(t, resultModel)
}

func TestFilePickerNoCallback(t *testing.T) {
	// Create a mock conversation manager
	manager := geppetto_conversation.NewManager()
	
	// Create a mock backend
	backend := &MockBackend{finished: true}

	// Create model without callback
	model := InitialModel(manager, backend)

	// Test that callback is nil
	assert.Nil(t, model.filePickerCallback)

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Simulate file selection - should use default saveToFile behavior
	resultModel, _ := model.handleFileSelection(tmpFile.Name())

	// Verify model returned successfully
	assert.NotNil(t, resultModel)
}
