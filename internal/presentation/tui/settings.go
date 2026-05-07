package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
)

// field indices for settings focus
const (
	fieldDefaultChain  = 0
	fieldKeyMode       = 1
	fieldGlobalKey     = 2
	fieldChainKeyStart = 3 // chain keys start here; index = fieldChainKeyStart + chainIdx
)

// defaultAmountIdx returns the field index for the default amount field
func (ss *SettingsScreen) defaultAmountIdx() int {
	if ss.cfg.App.KeyMode == "per_chain" {
		return fieldChainKeyStart + len(ss.chains)
	}
	return fieldGlobalKey + 1 // index 3 in global mode
}

// maxBatchWalletIdx returns the field index for the max batch wallet per tx field
func (ss *SettingsScreen) maxBatchWalletIdx() int {
	return ss.defaultAmountIdx() + 1
}

// SettingsScreen displays and edits application settings
type SettingsScreen struct {
	cfg       *config.AppConfig
	chains    []*config.ChainConfig // available chain configs (for per_chain mode)
	focusIdx  int
	inputMode bool   // true when actively typing in a text field
	inputBuf  string // current input buffer
	showKey   bool   // whether to show keys in plaintext
	width     int
}

// NewSettingsScreen creates a new settings screen
func NewSettingsScreen(cfg *config.AppConfig, chains []*config.ChainConfig) *SettingsScreen {
	return &SettingsScreen{
		cfg:    cfg,
		chains: chains,
	}
}

// maxField returns the maximum focus index for the current mode
func (ss *SettingsScreen) maxField() int {
	return ss.maxBatchWalletIdx()
}

// Update handles keyboard input for the settings screen
func (ss *SettingsScreen) Update(msg tea.Msg) (*SettingsScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// If actively typing in a field
		if ss.inputMode {
			switch key {
			case "enter":
				wasKeyField := ss.isKeyField(ss.focusIdx)
				ss.applyInput()
				ss.inputMode = false
				if wasKeyField {
					return ss, func() tea.Msg { return privateKeyChangedMsg{} }
				}
				return ss, nil
			case "esc":
				ss.inputMode = false
				ss.inputBuf = ""
				return ss, nil
			case "backspace":
				if len(ss.inputBuf) > 0 {
					ss.inputBuf = ss.inputBuf[:len(ss.inputBuf)-1]
				}
				return ss, nil
			default:
				if len(key) == 1 {
					c := key[0]
					// Amount field accepts digits and dots
					if ss.focusIdx == ss.defaultAmountIdx() {
						if (c >= '0' && c <= '9') || c == '.' {
							ss.inputBuf += key
						}
					} else if ss.focusIdx == ss.maxBatchWalletIdx() {
						// Batch wallet field accepts digits only
						if c >= '0' && c <= '9' {
							ss.inputBuf += key
						}
					} else {
						// Key fields accept hex characters and 0x prefix
						if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == 'x' {
							ss.inputBuf += key
						}
					}
				}
				return ss, nil
			}
		}

		// Not in input mode — navigation
		switch key {
		case "esc":
			return ss, func() tea.Msg { return goToMenuMsg{} }

		case "ctrl+s":
			return ss, ss.saveConfig()

		case "up", "k":
			if ss.focusIdx > 0 {
				ss.focusIdx--
			}

		case "down", "j":
			if ss.focusIdx < ss.maxField() {
				ss.focusIdx++
			}

		case "left", "h":
			if ss.focusIdx == fieldDefaultChain {
				ss.PrevChain()
			} else if ss.focusIdx == fieldKeyMode {
				ss.cfg.App.KeyMode = "global"
			}

		case "right", "l":
			if ss.focusIdx == fieldDefaultChain {
				ss.NextChain()
			} else if ss.focusIdx == fieldKeyMode {
				ss.cfg.App.KeyMode = "per_chain"
			}

		case "enter", " ":
			if ss.focusIdx == fieldDefaultChain {
				ss.NextChain()
			} else if ss.focusIdx == fieldKeyMode {
				// Toggle key mode
				if ss.cfg.App.KeyMode == "global" {
					ss.cfg.App.KeyMode = "per_chain"
				} else {
					ss.cfg.App.KeyMode = "global"
				}
				// Clamp focus
				if ss.focusIdx > ss.maxField() {
					ss.focusIdx = ss.maxField()
				}
			} else {
				// Start input mode for key fields
				ss.startInput()
			}

		case "t":
			ss.showKey = !ss.showKey
		}
	}

	return ss, nil
}

// startInput begins input mode for the focused field
func (ss *SettingsScreen) startInput() {
	// Skip selector fields (default chain, key mode) — they use left/right, not input mode
	if ss.focusIdx == fieldDefaultChain || ss.focusIdx == fieldKeyMode {
		return
	}

	// Default amount field
	if ss.focusIdx == ss.defaultAmountIdx() {
		ss.inputBuf = ss.cfg.App.DefaultAmount
		ss.inputMode = true
		return
	}

	// Max batch wallet per tx field
	if ss.focusIdx == ss.maxBatchWalletIdx() {
		ss.inputBuf = strconv.Itoa(ss.cfg.App.MaxBatchWalletPerTx)
		ss.inputMode = true
		return
	}

	// Global private key
	if ss.focusIdx == fieldGlobalKey && ss.cfg.App.KeyMode == "global" {
		ss.inputBuf = ss.cfg.App.SenderPrivateKey
		ss.inputMode = true
		return
	}

	// Per-chain keys
	if ss.cfg.App.KeyMode == "per_chain" && ss.focusIdx >= fieldChainKeyStart {
		chainIdx := ss.focusIdx - fieldChainKeyStart
		if chainIdx < len(ss.chains) {
			chainKey := ss.chains[chainIdx].Key
			ss.inputBuf = ss.cfg.App.ChainKeys[chainKey]
			ss.inputMode = true
		}
	}
}

// isKeyField returns true if the given field index is a private key field
func (ss *SettingsScreen) isKeyField(idx int) bool {
	if ss.cfg.App.KeyMode == "global" && idx == fieldGlobalKey {
		return true
	}
	if ss.cfg.App.KeyMode == "per_chain" && idx >= fieldChainKeyStart {
		return true
	}
	return false
}

// applyInput applies the current input buffer to the focused field
func (ss *SettingsScreen) applyInput() {
	// Default amount field
	if ss.focusIdx == ss.defaultAmountIdx() {
		ss.cfg.App.DefaultAmount = ss.inputBuf
		ss.inputBuf = ""
		return
	}

	// Max batch wallet per tx field
	if ss.focusIdx == ss.maxBatchWalletIdx() {
		val, err := strconv.Atoi(ss.inputBuf)
		if err == nil && val >= 1 {
			ss.cfg.App.MaxBatchWalletPerTx = val
		}
		ss.inputBuf = ""
		return
	}

	// Global private key
	if ss.focusIdx == fieldGlobalKey && ss.cfg.App.KeyMode == "global" {
		ss.cfg.App.SenderPrivateKey = ss.inputBuf
		ss.inputBuf = ""
		return
	}

	// Per-chain keys
	if ss.cfg.App.KeyMode == "per_chain" && ss.focusIdx >= fieldChainKeyStart {
		chainIdx := ss.focusIdx - fieldChainKeyStart
		if chainIdx < len(ss.chains) {
			chainKey := ss.chains[chainIdx].Key
			if ss.cfg.App.ChainKeys == nil {
				ss.cfg.App.ChainKeys = map[string]string{}
			}
			ss.cfg.App.ChainKeys[chainKey] = ss.inputBuf
			ss.inputBuf = ""
		}
	}
}

// saveConfig saves the current config to disk
func (ss *SettingsScreen) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := ss.cfg.Save(); err != nil {
			return settingsErrMsg(err.Error())
		}
		return settingsSavedMsg2{}
	}
}

// View renders the settings screen
func (ss *SettingsScreen) View(width int) string {
	ss.width = width
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("  Settings") + "\n\n")

	// ── Default Chain ──
	sb.WriteString(ss.renderDefaultChain() + "\n\n")

	// ── Key Mode ──
	sb.WriteString(ss.renderKeyMode() + "\n\n")

	// ── Key Input ──
	if ss.cfg.App.KeyMode == "global" {
		sb.WriteString(ss.renderGlobalKey() + "\n\n")
	} else {
		sb.WriteString(ss.renderChainKeys() + "\n\n")
	}

	// ── Default Amount ──
	sb.WriteString(ss.renderDefaultAmount() + "\n\n")

	// ── Max Batch Wallet Per Tx ──
	sb.WriteString(ss.renderMaxBatchWallet() + "\n")

	return sb.String()
}

func (ss *SettingsScreen) renderDefaultChain() string {
	label := "  Default Chain:  "
	focused := ss.focusIdx == fieldDefaultChain

	// Find display name for the stored chain key
	chainVal := "(none)"
	for _, c := range ss.chains {
		if c.Key == ss.cfg.App.DefaultChain {
			chainVal = c.Name
			break
		}
	}

	var sb strings.Builder
	if focused {
		sb.WriteString(Highlight(label+chainVal) + MutedStyle.Render("  [←→ to change]"))
	} else {
		sb.WriteString("    " + label + ToolLabelStyle.Render(chainVal))
	}

	return sb.String()
}

func (ss *SettingsScreen) renderKeyMode() string {
	label := "  Key Mode:      "
	focused := ss.focusIdx == fieldKeyMode

	globalBtn := "  Global  "
	perChainBtn := "  Per Chain  "

	if ss.cfg.App.KeyMode == "global" {
		if focused {
			globalBtn = SelectedStyle.Render(globalBtn)
		} else {
			globalBtn = ActiveTabStyle.Render(globalBtn)
		}
		perChainBtn = InactiveTabStyle.Render(perChainBtn)
	} else {
		globalBtn = InactiveTabStyle.Render(globalBtn)
		if focused {
			perChainBtn = SelectedStyle.Render(perChainBtn)
		} else {
			perChainBtn = ActiveTabStyle.Render(perChainBtn)
		}
	}

	return label + globalBtn + " / " + perChainBtn
}

func (ss *SettingsScreen) renderGlobalKey() string {
	focused := ss.focusIdx == fieldGlobalKey && ss.inputMode
	editFocused := ss.focusIdx == fieldGlobalKey && !ss.inputMode

	label := "  Sender Key:    "

	var val string
	if focused {
		// Show input buffer with cursor
		val = ss.inputBuf + "█"
		if !ss.showKey && len(val) > 1 {
			val = maskString(ss.inputBuf) + "█"
		}
		return Highlight(label + val)
	} else if editFocused {
		val = ss.displayKeyValue(ss.cfg.App.SenderPrivateKey)
		return KeyStyle.Render(" > ") + label + ToolLabelStyle.Render(val) +
			MutedStyle.Render("  [enter to edit]")
	}

	val = ss.displayKeyValue(ss.cfg.App.SenderPrivateKey)
	return "    " + label + MutedStyle.Render(val)
}

func (ss *SettingsScreen) renderChainKeys() string {
	var sb strings.Builder

	sb.WriteString(SectionHeaderStyle.Width(ss.width).Render(" PER-CHAIN PRIVATE KEYS") + "\n")

	if len(ss.chains) == 0 {
		sb.WriteString(MutedStyle.Render("  No chains configured. Add chain YAML files in configs/chains/") + "\n")
		return sb.String()
	}

	for i, chain := range ss.chains {
		fieldIdx := fieldChainKeyStart + i
		focused := ss.focusIdx == fieldIdx && ss.inputMode
		editFocused := ss.focusIdx == fieldIdx && !ss.inputMode

		chainLabel := fmt.Sprintf("  %-12s", chain.Key+":")
		keyVal := ss.cfg.App.ChainKeys[chain.Key]

		var line string
		if focused {
			displayVal := ss.inputBuf + "█"
			if !ss.showKey && len(displayVal) > 1 {
				displayVal = maskString(ss.inputBuf) + "█"
			}
			line = Highlight(chainLabel + displayVal)
		} else if editFocused {
			displayVal := ss.displayKeyValue(keyVal)
			line = KeyStyle.Render(" > ") + chainLabel + ToolLabelStyle.Render(displayVal) +
				MutedStyle.Render("  [enter to edit]")
		} else {
			displayVal := ss.displayKeyValue(keyVal)
			line = "    " + chainLabel + MutedStyle.Render(displayVal)
		}

		sb.WriteString(line + "\n")
	}

	return sb.String()
}

// NextChain cycles to the next available chain
func (ss *SettingsScreen) NextChain() {
	if len(ss.chains) == 0 {
		ss.cfg.App.DefaultChain = ""
		return
	}
	// Find current index
	currentIdx := -1
	for i, c := range ss.chains {
		if c.Key == ss.cfg.App.DefaultChain {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(ss.chains)
	ss.cfg.App.DefaultChain = ss.chains[nextIdx].Key
}

// PrevChain cycles to the previous available chain
func (ss *SettingsScreen) PrevChain() {
	if len(ss.chains) == 0 {
		ss.cfg.App.DefaultChain = ""
		return
	}
	// Find current index
	currentIdx := -1
	for i, c := range ss.chains {
		if c.Key == ss.cfg.App.DefaultChain {
			currentIdx = i
			break
		}
	}
	prevIdx := (currentIdx - 1 + len(ss.chains)) % len(ss.chains)
	ss.cfg.App.DefaultChain = ss.chains[prevIdx].Key
}

func (ss *SettingsScreen) renderDefaultAmount() string {
	focused := ss.focusIdx == ss.defaultAmountIdx() && ss.inputMode
	editFocused := ss.focusIdx == ss.defaultAmountIdx() && !ss.inputMode

	label := "  Default Amount: "

	if focused {
		return Highlight(label + ss.inputBuf + "█")
	} else if editFocused {
		return KeyStyle.Render(" > ") + label + ToolLabelStyle.Render(ss.cfg.App.DefaultAmount) +
			MutedStyle.Render("  [enter to edit]")
	}

	return "    " + label + MutedStyle.Render(ss.cfg.App.DefaultAmount)
}

func (ss *SettingsScreen) renderMaxBatchWallet() string {
	focused := ss.focusIdx == ss.maxBatchWalletIdx() && ss.inputMode
	editFocused := ss.focusIdx == ss.maxBatchWalletIdx() && !ss.inputMode

	label := "  Max Batch/Tx:   "
	val := strconv.Itoa(ss.cfg.App.MaxBatchWalletPerTx)

	if focused {
		return Highlight(label + ss.inputBuf + "█")
	} else if editFocused {
		return KeyStyle.Render(" > ") + label + ToolLabelStyle.Render(val) +
			MutedStyle.Render("  [enter to edit]  wallets per transaction")
	}

	return "    " + label + MutedStyle.Render(val) + MutedStyle.Render("  wallets per transaction")
}

// displayKeyValue returns a masked or plaintext key value
func (ss *SettingsScreen) displayKeyValue(val string) string {
	if val == "" {
		return "[not set]"
	}
	if ss.showKey {
		return val
	}
	return maskString(val)
}

// maskString masks a string with bullets, showing first 6 and last 4 chars
func maskString(s string) string {
	if len(s) <= 10 {
		return strings.Repeat("•", len(s))
	}
	return s[:6] + strings.Repeat("•", len(s)-10) + s[len(s)-4:]
}

// Messages for settings
type settingsSavedMsg2 struct{}
type settingsErrMsg string
