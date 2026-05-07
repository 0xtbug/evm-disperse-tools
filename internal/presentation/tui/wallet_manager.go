package tui

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ethereum/go-ethereum/crypto"
)

// walletEntry represents a single generated wallet
type walletEntry struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
}

// walletFile is the JSON structure saved to disk
type walletFile struct {
	Wallets []walletEntry `json:"wallets"`
}

// WalletManagerModel manages wallet generation and saving
type WalletManagerModel struct {
	// Input
	numInput  string
	nameInput string // name for the wallet list
	focused   int    // 0 = num input, 1 = name input
	inputMode bool   // true when actively typing in a field

	// State
	wallets []walletEntry

	// Config
	walletsDir        string // configs/wallets/ directory
	maxWalletGenerate int
	width             int
}

// NewWalletManagerModel creates a new wallet manager
func NewWalletManagerModel(maxWalletGenerate int) *WalletManagerModel {
	if maxWalletGenerate < 1 {
		maxWalletGenerate = 1000
	}
	return &WalletManagerModel{
		numInput:          "1",
		nameInput:         "default",
		focused:           0,
		walletsDir:        filepath.Join("configs", "wallets"),
		maxWalletGenerate: maxWalletGenerate,
	}
}

// walletGeneratedMsg is sent when wallets are generated
type walletGeneratedMsg struct {
	wallets []walletEntry
}

// walletSavedMsg is sent when wallets are saved
type walletSavedMsg struct {
	path string
}

// Init initializes the wallet manager
func (wm *WalletManagerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the wallet manager
func (wm *WalletManagerModel) Update(msg tea.Msg) (*WalletManagerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Input mode: handle text editing in the focused field
		if wm.inputMode {
			switch key {
			case "enter":
				wm.inputMode = false
				return wm, nil
			case "esc":
				wm.inputMode = false
				return wm, nil
			case "tab":
				wm.inputMode = false
				if wm.focused == 0 {
					wm.focused = 1
				} else {
					wm.focused = 0
				}
				return wm, nil
			case "backspace":
				if wm.focused == 0 && len(wm.numInput) > 0 {
					wm.numInput = wm.numInput[:len(wm.numInput)-1]
				} else if wm.focused == 1 && len(wm.nameInput) > 0 {
					wm.nameInput = wm.nameInput[:len(wm.nameInput)-1]
				}
				return wm, nil
			default:
				if len(key) == 1 {
					if wm.focused == 0 && key[0] >= '0' && key[0] <= '9' && len(wm.numInput) < 4 {
						wm.numInput += key
					} else if wm.focused == 1 {
						c := key[0]
						if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') && len(wm.nameInput) < 30 {
							wm.nameInput += key
						}
					}
				}
				return wm, nil
			}
		}

		// Navigation mode — shortcuts always work here
		switch key {
		case "tab":
			if wm.focused == 0 {
				wm.focused = 1
			} else {
				wm.focused = 0
			}
			return wm, nil
		case "enter", " ":
			wm.inputMode = true
			return wm, nil
		case "g", "G":
			return wm, wm.generateWallets()
		case "s", "S":
			return wm, wm.saveWallets()
		}

	case walletGeneratedMsg:
		wm.wallets = msg.wallets
		return wm, nil

	case walletSavedMsg:
		return wm, nil
	}

	return wm, nil
}

// generateWallets creates the requested number of wallets
func (wm *WalletManagerModel) generateWallets() tea.Cmd {
	return func() tea.Msg {
		num := 1
		if wm.numInput != "" {
			n, err := strconv.Atoi(wm.numInput)
			if err == nil && n > 0 {
				num = n
			}
		}

		if num > wm.maxWalletGenerate {
			num = wm.maxWalletGenerate
		}

		wallets := make([]walletEntry, num)
		for i := 0; i < num; i++ {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				continue
			}

			privateKeyBytes := crypto.FromECDSA(privateKey)
			publicKey := privateKey.Public().(*ecdsa.PublicKey)
			address := crypto.PubkeyToAddress(*publicKey)

			wallets[i] = walletEntry{
				Address:    address.Hex(),
				PrivateKey: fmt.Sprintf("%x", privateKeyBytes),
			}
		}

		return walletGeneratedMsg{wallets: wallets}
	}
}

// saveWallets writes the wallets to a named list file
func (wm *WalletManagerModel) saveWallets() tea.Cmd {
	return func() tea.Msg {
		if len(wm.wallets) == 0 {
			return walletSavedMsg{path: "error: no wallets to save — press G first"}
		}

		name := wm.nameInput
		if name == "" {
			name = "default"
		}

		// Ensure wallets directory exists
		if err := os.MkdirAll(wm.walletsDir, 0755); err != nil {
			return walletSavedMsg{path: fmt.Sprintf("error: %v", err)}
		}

		// Build the wallet file data
		wf := walletFile{Wallets: wm.wallets}
		data, err := json.MarshalIndent(wf, "", "  ")
		if err != nil {
			return walletSavedMsg{path: fmt.Sprintf("error: %v", err)}
		}

		// Save to configs/wallets/<name>.json only
		namedPath := filepath.Join(wm.walletsDir, name+".json")
		if err := os.WriteFile(namedPath, data, 0644); err != nil {
			return walletSavedMsg{path: fmt.Sprintf("error: %v", err)}
		}

		return walletSavedMsg{path: namedPath}
	}
}

// View renders the wallet manager
func (wm *WalletManagerModel) View(width int) string {
	wm.width = width
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("  Wallet Manager") + "\n\n")

	// Number of wallets input
	if wm.focused == 0 {
		if wm.inputMode {
			sb.WriteString(Highlight("  > Number of wallets: "+wm.numInput+"█") + "\n")
		} else {
			sb.WriteString(KeyStyle.Render(" > ") + "Number of wallets: " + ToolLabelStyle.Render(wm.numInput) + MutedStyle.Render("  [enter to edit]") + "\n")
		}
	} else {
		sb.WriteString("    Number of wallets: " + ToolLabelStyle.Render(wm.numInput) + "\n")
	}

	// List name input
	if wm.focused == 1 {
		if wm.inputMode {
			sb.WriteString(Highlight("  > List name:        "+wm.nameInput+"█") + "\n")
		} else {
			sb.WriteString(KeyStyle.Render(" > ") + "List name:        " + ToolLabelStyle.Render(wm.nameInput) + MutedStyle.Render("  [enter to edit]") + "\n")
		}
	} else {
		sb.WriteString("    List name:        " + ToolLabelStyle.Render(wm.nameInput) + "\n")
	}

	sb.WriteString("\n")

	return sb.String()
}

// GetNumWallets returns the current number input as an integer
func (wm *WalletManagerModel) GetNumWallets() int {
	n, err := strconv.Atoi(wm.numInput)
	if err != nil || n <= 0 {
		return 1
	}
	if n > wm.maxWalletGenerate {
		return wm.maxWalletGenerate
	}
	return n
}
