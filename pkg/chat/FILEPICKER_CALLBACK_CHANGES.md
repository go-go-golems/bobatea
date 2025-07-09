# File Picker Callback Implementation

## Summary of Changes

The filepicker implementation in the bobatea chat package has been modified to support a callback mechanism for file selection. This change provides more flexibility for handling file selection actions.

## Changes Made

### 1. Removed Globbing (✓)
- Removed `filepicker.WithGlobPattern("*.{json,yaml,yml}")` from the filepicker initialization in `InitialModel` function
- The filepicker now shows all files without filtering by file extension

### 2. Added Callback Mechanism (✓)
- Added `filePickerCallback func(path string) error` field to the `model` struct
- Added `WithFilePickerCallback(callback func(path string) error) ModelOption` function to configure the callback
- The callback is called when a file is selected or created in the filepicker

### 3. Modified Save Handling (✓)
- Created new `handleFileSelection(path string) (tea.Model, tea.Cmd)` method that:
  - Uses the callback if provided
  - Falls back to the original `saveToFile` behavior if no callback is set
  - Handles errors from the callback appropriately
- Updated `SelectFileMsg` handler to use `handleFileSelection` instead of `saveToFile`

## API Usage

### Setting up a callback:
```go
// Define a callback that writes "hello" to the selected file
callback := func(path string) error {
    return os.WriteFile(path, []byte("hello"), 0644)
}

// Create model with callback
model := InitialModel(manager, backend, WithFilePickerCallback(callback))
```

### Default behavior (no callback):
```go
// Create model without callback - uses original saveToFile behavior
model := InitialModel(manager, backend)
```

## Features Maintained

- All existing filepicker features remain intact:
  - File preview
  - Detailed view
  - Hidden file visibility control
  - Directory navigation
  - File creation capabilities
- Backward compatibility maintained - existing code continues to work without changes
- Error handling preserves original behavior
- State management remains consistent

## Testing

- Added comprehensive tests in `callback_test.go`:
  - `TestFilePickerCallback`: Tests successful callback execution
  - `TestFilePickerCallbackWithError`: Tests error handling in callbacks
  - `TestFilePickerNoCallback`: Tests default behavior without callback
- All tests pass successfully

## Example Usage

The callback system enables flexible file handling scenarios:

1. **Write specific content to selected files**
2. **Process files before saving**
3. **Integrate with external systems**
4. **Custom file validation**
5. **Multi-step file operations**

The implementation provides a clean, flexible interface while maintaining all existing functionality and backward compatibility.
