package tui

import (
	"fmt"
	"strings"

	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/storage"
)

// DisperseFormScreen represents a form for disperse operations
type DisperseFormScreen struct {
	chainOptions     []string
	selectedChainIdx int
	chain            string

	// Wallet list selection
	walletLists      []storage.WalletListInfo
	selectedListIdx  int
	selectedListName string
	selectedAddrs    []string // loaded addresses from selected list

	// Text input fields
	token  string
	amount string

	// Input state
	mode      string // "Native" or "ERC20"
	focused   int    // 0: chain, 1: wallet list, 2: token (erc20 only), 3: amount
	inputMode bool
	inputBuf  string
}

// NewDisperseFormScreen creates a new disperse form
func NewDisperseFormScreen(mode string, chainOptions []string, walletLists []storage.WalletListInfo, defaultAmount string, defaultChain string) *DisperseFormScreen {
	selectedIdx := 0
	selectedChain := ""
	if len(chainOptions) > 0 {
		// Find the default chain if specified
		if defaultChain != "" {
			for i, name := range chainOptions {
				if strings.EqualFold(name, defaultChain) {
					selectedIdx = i
					break
				}
			}
		}
		selectedChain = chainOptions[selectedIdx]
	}

	selectedListName := ""
	selectedAddrs := []string{}
	if len(walletLists) > 0 {
		selectedListName = walletLists[0].Name
		// Load addresses from the first list
		addrs, _ := storage.LoadAddressesFromWalletManager(walletLists[0].Path)
		selectedAddrs = addrs
	}

	amount := defaultAmount
	if amount == "" {
		amount = "0.01"
	}

	return &DisperseFormScreen{
		chainOptions:     chainOptions,
		selectedChainIdx: selectedIdx,
		chain:            selectedChain,
		walletLists:      walletLists,
		selectedListIdx:  0,
		selectedListName: selectedListName,
		selectedAddrs:    selectedAddrs,
		token:            "",
		amount:           amount,
		mode:             mode,
		focused:          0,
	}
}

// maxFieldIdx returns the max field index
func (dfs *DisperseFormScreen) maxFieldIdx() int {
	if dfs.mode == "ERC20" {
		return 3 // chain, list, token, amount
	}
	return 2 // chain, list, amount
}

// Update handles keyboard input for the form
func (dfs *DisperseFormScreen) Update(key string) bool {
	// If in input mode, handle text editing
	if dfs.inputMode {
		switch key {
		case "enter":
			dfs.applyInput()
			dfs.inputMode = false
			return true
		case "esc":
			dfs.inputMode = false
			dfs.inputBuf = ""
			return true
		case "backspace":
			if len(dfs.inputBuf) > 0 {
				dfs.inputBuf = dfs.inputBuf[:len(dfs.inputBuf)-1]
			}
			return true
		default:
			if len(key) == 1 {
				dfs.inputBuf += key
			}
			return true
		}
	}

	// Navigation mode
	switch key {
	case "tab":
		dfs.focused++
		if dfs.focused > dfs.maxFieldIdx() {
			dfs.focused = 0
		}
		return true
	case "shift+tab":
		dfs.focused--
		if dfs.focused < 0 {
			dfs.focused = dfs.maxFieldIdx()
		}
		return true
	case "up":
		if dfs.focused == 0 {
			dfs.PrevChain()
		} else if dfs.focused == 1 {
			dfs.PrevWalletList()
		}
		return true
	case "down":
		if dfs.focused == 0 {
			dfs.NextChain()
		} else if dfs.focused == 1 {
			dfs.NextWalletList()
		}
		return true
	case "enter":
		// Start input mode for text fields
		dfs.startInput()
		return true
	}

	return false
}

// startInput enters input mode for the focused field
func (dfs *DisperseFormScreen) startInput() {
	if dfs.mode == "ERC20" {
		switch dfs.focused {
		case 2: // token
			dfs.inputBuf = dfs.token
			dfs.inputMode = true
		case 3: // amount
			dfs.inputBuf = dfs.amount
			dfs.inputMode = true
		}
	} else {
		switch dfs.focused {
		case 2: // amount
			dfs.inputBuf = dfs.amount
			dfs.inputMode = true
		}
	}
}

// applyInput saves the input buffer to the focused field
func (dfs *DisperseFormScreen) applyInput() {
	if dfs.mode == "ERC20" {
		switch dfs.focused {
		case 2:
			dfs.token = dfs.inputBuf
		case 3:
			dfs.amount = dfs.inputBuf
		}
	} else {
		switch dfs.focused {
		case 2:
			dfs.amount = dfs.inputBuf
		}
	}
	dfs.inputBuf = ""
}

// NextWalletList moves to the next wallet list
func (dfs *DisperseFormScreen) NextWalletList() {
	if len(dfs.walletLists) == 0 {
		return
	}
	dfs.selectedListIdx = (dfs.selectedListIdx + 1) % len(dfs.walletLists)
	dfs.selectedListName = dfs.walletLists[dfs.selectedListIdx].Name
	addrs, _ := storage.LoadAddressesFromWalletManager(dfs.walletLists[dfs.selectedListIdx].Path)
	dfs.selectedAddrs = addrs
}

// PrevWalletList moves to the previous wallet list
func (dfs *DisperseFormScreen) PrevWalletList() {
	if len(dfs.walletLists) == 0 {
		return
	}
	dfs.selectedListIdx = (dfs.selectedListIdx - 1 + len(dfs.walletLists)) % len(dfs.walletLists)
	dfs.selectedListName = dfs.walletLists[dfs.selectedListIdx].Name
	addrs, _ := storage.LoadAddressesFromWalletManager(dfs.walletLists[dfs.selectedListIdx].Path)
	dfs.selectedAddrs = addrs
}

// NextChain moves to the next chain option
func (dfs *DisperseFormScreen) NextChain() {
	if len(dfs.chainOptions) == 0 {
		return
	}
	dfs.selectedChainIdx = (dfs.selectedChainIdx + 1) % len(dfs.chainOptions)
	dfs.chain = dfs.chainOptions[dfs.selectedChainIdx]
}

// PrevChain moves to the previous chain option
func (dfs *DisperseFormScreen) PrevChain() {
	if len(dfs.chainOptions) == 0 {
		return
	}
	dfs.selectedChainIdx = (dfs.selectedChainIdx - 1 + len(dfs.chainOptions)) % len(dfs.chainOptions)
	dfs.chain = dfs.chainOptions[dfs.selectedChainIdx]
}

// View renders the form with responsive width
func (dfs *DisperseFormScreen) View(width int) string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("  Disperse "+dfs.mode) + "\n\n")

	// ── Chain selector ──
	if dfs.focused == 0 {
		sb.WriteString(Highlight(fmt.Sprintf("> Chain: %s  [↑↓]", dfs.chain)) + "\n")
	} else {
		sb.WriteString("    Chain: " + ToolLabelStyle.Render(dfs.chain) + "\n")
	}

	// ── Wallet list selector ──
	if dfs.focused == 1 {
		listLabel := fmt.Sprintf("> Wallet List: %s  [↑↓]  (%d addresses)", dfs.selectedListName, len(dfs.selectedAddrs))
		sb.WriteString(Highlight(listLabel) + "\n")
	} else {
		listLabel := fmt.Sprintf("  Wallet List: %s  (%d addresses)", dfs.selectedListName, len(dfs.selectedAddrs))
		sb.WriteString("    " + ToolLabelStyle.Render(listLabel) + "\n")
	}

	// ── Token field (ERC20 only) ──
	tokenFieldIdx := -1
	if dfs.mode == "ERC20" {
		tokenFieldIdx = 2
		if dfs.focused == tokenFieldIdx {
			if dfs.inputMode && dfs.focused == tokenFieldIdx {
				sb.WriteString(Highlight("> Token: "+dfs.inputBuf+"█") + "\n")
			} else {
				tokenVal := dfs.token
				if tokenVal == "" {
					tokenVal = MutedStyle.Render("[enter to set]")
				}
				sb.WriteString(KeyStyle.Render(" > ") + "Token: " + ToolLabelStyle.Render(tokenVal) + "\n")
			}
		} else {
			tokenVal := dfs.token
			if tokenVal == "" {
				tokenVal = MutedStyle.Render("[not set]")
			}
			sb.WriteString("    Token: " + MutedStyle.Render(tokenVal) + "\n")
		}
	}

	// ── Amount field ──
	amountFieldIdx := 3
	if dfs.mode == "Native" {
		amountFieldIdx = 2
	}
	if dfs.focused == amountFieldIdx {
		if dfs.inputMode && dfs.focused == amountFieldIdx {
			sb.WriteString(Highlight("> Amount: "+dfs.inputBuf+"█") + "\n")
		} else {
			sb.WriteString(KeyStyle.Render(" > ") + "Amount: " + ToolLabelStyle.Render(dfs.amount) + MutedStyle.Render("  [enter to edit]") + "\n")
		}
	} else {
		sb.WriteString("    Amount: " + MutedStyle.Render(dfs.amount) + "\n")
	}

	return sb.String()
}

// GetSelectedChainIdx returns the selected chain index
func (dfs *DisperseFormScreen) GetSelectedChainIdx() int {
	return dfs.selectedChainIdx
}

// GetRecipients returns the selected wallet list addresses
func (dfs *DisperseFormScreen) GetRecipients() []string {
	return dfs.selectedAddrs
}

// GetAmount returns the amount
func (dfs *DisperseFormScreen) GetAmount() string {
	return dfs.amount
}

// GetToken returns the token address
func (dfs *DisperseFormScreen) GetToken() string {
	return dfs.token
}

// UpdateDefaultChain updates the selected chain to match the given default chain name
func (dfs *DisperseFormScreen) UpdateDefaultChain(chainName string) {
	if chainName == "" {
		return
	}
	for i, name := range dfs.chainOptions {
		if strings.EqualFold(name, chainName) {
			dfs.selectedChainIdx = i
			dfs.chain = dfs.chainOptions[i]
			return
		}
	}
}

// UpdateDefaultAmount updates the amount field
func (dfs *DisperseFormScreen) UpdateDefaultAmount(amount string) {
	if amount != "" {
		dfs.amount = amount
	}
}

// UpdateWalletLists refreshes the wallet list selection and reloads addresses.
// Always reloads addresses to ensure the form has the latest data (e.g. after
// the user generates and saves new wallets to an existing list name).
func (dfs *DisperseFormScreen) UpdateWalletLists(lists []storage.WalletListInfo) {
	dfs.walletLists = lists
	found := false
	for i, l := range lists {
		if l.Name == dfs.selectedListName {
			dfs.selectedListIdx = i
			found = true
			break
		}
	}
	if !found && len(lists) > 0 {
		dfs.selectedListIdx = 0
		dfs.selectedListName = lists[0].Name
	}
	// Always reload addresses from the selected list to pick up any changes
	if len(dfs.walletLists) > 0 && dfs.selectedListIdx < len(dfs.walletLists) {
		addrs, _ := storage.LoadAddressesFromWalletManager(dfs.walletLists[dfs.selectedListIdx].Path)
		dfs.selectedAddrs = addrs
	}
}

// IsValid checks if form is valid
func (dfs *DisperseFormScreen) IsValid() bool {
	if dfs.chain == "" || dfs.amount == "" || len(dfs.selectedAddrs) == 0 {
		return false
	}
	if dfs.mode == "ERC20" && dfs.token == "" {
		return false
	}
	return true
}
