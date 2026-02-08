## 1. Add CLI flag and parameter handling

- [x] 1.1 Add `PrintAndExit` field to `Options` struct in `pkg/commands/root.go`
- [x] 1.2 Add `-p` / `--print` flag binding in `Options.AddPFlags()`
- [x] 1.3 Modify `RunE` to handle three modes based on args and -p flag:
  - No args, no -p: normal interactive mode
  - Has args, no -p: interactive mode with initial prompt
  - Has args, has -p: non-interactive single-query mode
- [x] 1.4 Pass `initialPrompt` and `autoExitAfterResponse` to ChatUI in all code paths

## 2. Extend ChatUI Options and state

- [x] 2.1 Add `InitialPrompt` field to `chat.Options` struct in `pkg/ui/chat/ui.go`
- [x] 2.2 Add `AutoExitAfterResponse` field to `chat.Options` struct
- [x] 2.3 Add `initialPrompt` field to `ChatUI` struct
- [x] 2.4 Add `autoExitAfterResponse` field to `ChatUI` struct
- [x] 2.5 Update `NewChatUI()` to initialize new fields from options

## 3. Implement initial prompt sending

- [x] 3.1 Modify `Init()` in `pkg/ui/chat/ui.go` to send initial prompt if set
- [x] 3.2 Ensure initial prompt is sent only after `newSession` completes (use `tea.Sequence`)
- [x] 3.3 Test that initial prompt appears as user message in UI

## 4. Implement auto-exit after response

- [x] 4.1 Modify `Update()` in `pkg/ui/chat/ui.go` to handle `PromptResponse`
- [x] 4.2 Add logic to return `tea.Quit` when `PromptResponse` received and `autoExitAfterResponse` is true
- [x] 4.3 Ensure `vp.Flush()` is called before quitting to output all cached messages

## 5. Testing

- [x] 5.1 Test `nfa` - normal interactive mode unchanged
- [x] 5.2 Test `nfa '问题'` - interactive mode with initial prompt
- [x] 5.3 Test `nfa '问题' -p` - non-interactive single-query mode
- [x] 5.4 Test `nfa '问题1' '问题2'` - only first argument used
- [x] 5.5 Test `nfa -p` - interactive mode (no auto-exit without question)
- [x] 5.6 Verify output consistency between interactive and non-interactive modes
