package tui

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
)

// Gas estimation constants
const (
	defaultGasPerRecipient = 36800
	baseGasPerTx           = 50000
)

// FeeCalculatorScreen displays a fee estimation summary for bulk native transfers
type FeeCalculatorScreen struct {
	chains           []*config.ChainConfig
	selectedChainIdx int
	appCfg           *config.AppConfig

	// Input fields
	recipients      string // number of recipients
	amountPerWallet string // amount per wallet in native token

	// External state (updated from AppModel)
	gasPriceGwei string // current gas price like "125.686"

	// Input state
	focused   int // 0: chain, 1: recipients, 2: amount
	inputMode bool
	inputBuf  string
	width     int
}

// NewFeeCalculatorScreen creates a new fee calculator screen
func NewFeeCalculatorScreen(chains []*config.ChainConfig, appCfg *config.AppConfig, defaultChainName string) *FeeCalculatorScreen {
	selectedIdx := 0
	if len(chains) > 0 && defaultChainName != "" {
		for i, c := range chains {
			if strings.EqualFold(c.Name, defaultChainName) {
				selectedIdx = i
				break
			}
		}
	}

	return &FeeCalculatorScreen{
		chains:           chains,
		selectedChainIdx: selectedIdx,
		appCfg:           appCfg,
		recipients:       "10000",
		amountPerWallet:  "0.000000000001",
		focused:          0,
	}
}

// maxFieldIdx returns the max field index
func (fc *FeeCalculatorScreen) maxFieldIdx() int {
	return 2
}

// Update handles keyboard input for the calculator
func (fc *FeeCalculatorScreen) Update(key string) bool {
	if fc.inputMode {
		switch key {
		case "enter":
			fc.applyInput()
			fc.inputMode = false
			return true
		case "esc":
			fc.inputMode = false
			fc.inputBuf = ""
			return true
		case "backspace":
			if len(fc.inputBuf) > 0 {
				fc.inputBuf = fc.inputBuf[:len(fc.inputBuf)-1]
			}
			return true
		default:
			if len(key) == 1 {
				c := key[0]
				if fc.focused == 1 { // recipients — digits only
					if c >= '0' && c <= '9' {
						fc.inputBuf += key
					}
				} else if fc.focused == 2 { // amount — digits and dots
					if (c >= '0' && c <= '9') || c == '.' {
						fc.inputBuf += key
					}
				}
			}
			return true
		}
	}

	switch key {
	case "tab":
		fc.focused++
		if fc.focused > fc.maxFieldIdx() {
			fc.focused = 0
		}
		return true
	case "shift+tab":
		fc.focused--
		if fc.focused < 0 {
			fc.focused = fc.maxFieldIdx()
		}
		return true
	case "up":
		if fc.focused == 0 {
			fc.PrevChain()
		}
		return true
	case "down":
		if fc.focused == 0 {
			fc.NextChain()
		}
		return true
	case "enter":
		if fc.focused > 0 {
			fc.startInput()
		}
		return true
	}
	return false
}

// startInput enters input mode for the focused field
func (fc *FeeCalculatorScreen) startInput() {
	switch fc.focused {
	case 1:
		fc.inputBuf = fc.recipients
		fc.inputMode = true
	case 2:
		fc.inputBuf = fc.amountPerWallet
		fc.inputMode = true
	}
}

// applyInput saves the input buffer to the focused field
func (fc *FeeCalculatorScreen) applyInput() {
	switch fc.focused {
	case 1:
		fc.recipients = fc.inputBuf
	case 2:
		fc.amountPerWallet = fc.inputBuf
	}
	fc.inputBuf = ""
}

// NextChain moves to the next chain
func (fc *FeeCalculatorScreen) NextChain() {
	if len(fc.chains) == 0 {
		return
	}
	fc.selectedChainIdx = (fc.selectedChainIdx + 1) % len(fc.chains)
}

// PrevChain moves to the previous chain
func (fc *FeeCalculatorScreen) PrevChain() {
	if len(fc.chains) == 0 {
		return
	}
	fc.selectedChainIdx = (fc.selectedChainIdx - 1 + len(fc.chains)) % len(fc.chains)
}

// SetGasPrice updates the gas price from the RPC health check
func (fc *FeeCalculatorScreen) SetGasPrice(gasPriceDisplay string) {
	// Strip " Gwei" suffix if present
	fc.gasPriceGwei = strings.TrimSuffix(strings.TrimSpace(gasPriceDisplay), "Gwei")
	fc.gasPriceGwei = strings.TrimSpace(fc.gasPriceGwei)
}

// getCurrentChain returns the currently selected chain config
func (fc *FeeCalculatorScreen) getCurrentChain() *config.ChainConfig {
	if len(fc.chains) == 0 || fc.selectedChainIdx >= len(fc.chains) {
		return nil
	}
	return fc.chains[fc.selectedChainIdx]
}

// GetSelectedChainIdx returns the selected chain index
func (fc *FeeCalculatorScreen) GetSelectedChainIdx() int {
	return fc.selectedChainIdx
}

// View renders the fee calculator screen
func (fc *FeeCalculatorScreen) View(width int) string {
	fc.width = width
	var sb strings.Builder

	chain := fc.getCurrentChain()
	if chain == nil {
		sb.WriteString(MutedStyle.Render("  No chains configured"))
		return sb.String()
	}
	nativeToken := chain.GetNativeToken()

	sb.WriteString(TitleStyle.Render("  Fee Calculator") + "\n\n")

	// ── Chain selector ──
	if fc.focused == 0 {
		sb.WriteString(Highlight(fmt.Sprintf("> Chain: %s  [↑↓]", chain.Name)) + "\n")
	} else {
		sb.WriteString("    Chain: " + ToolLabelStyle.Render(chain.Name) + "\n")
	}

	// ── Recipients field ──
	if fc.focused == 1 {
		if fc.inputMode {
			sb.WriteString(Highlight("> Recipients: "+fc.inputBuf+"█") + "\n")
		} else {
			sb.WriteString(KeyStyle.Render(" > ") + "Recipients: " + ToolLabelStyle.Render(fc.recipients) +
				MutedStyle.Render("  [enter to edit]") + "\n")
		}
	} else {
		sb.WriteString("    Recipients: " + MutedStyle.Render(fc.recipients) + "\n")
	}

	// ── Amount per wallet field ──
	if fc.focused == 2 {
		if fc.inputMode {
			sb.WriteString(Highlight("> Amount/Wallet: "+fc.inputBuf+"█") + "\n")
		} else {
			sb.WriteString(KeyStyle.Render(" > ") + "Amount/Wallet: " + ToolLabelStyle.Render(fc.amountPerWallet) +
				MutedStyle.Render("  [enter to edit]") + "\n")
		}
	} else {
		sb.WriteString("    Amount/Wallet: " + MutedStyle.Render(fc.amountPerWallet) + "\n")
	}

	sb.WriteString("\n")

	// ── Summary ──
	sb.WriteString(fc.renderSummary(nativeToken))

	return sb.String()
}

// renderSummary computes and renders the fee estimation summary
func (fc *FeeCalculatorScreen) renderSummary(nativeToken string) string {
	var sb strings.Builder

	maxBatch := fc.appCfg.App.MaxBatchWalletPerTx
	if maxBatch < 1 {
		maxBatch = 250
	}

	// Parse inputs
	totalRecipients := parseInt(fc.recipients)
	amountPerWalletStr := fc.amountPerWallet
	if amountPerWalletStr == "" {
		amountPerWalletStr = "0"
	}

	// Parse amount per wallet
	amountPerWallet, _, _ := big.ParseFloat(amountPerWalletStr, 10, 256, big.ToNearestEven)
	if amountPerWallet == nil {
		amountPerWallet = big.NewFloat(0)
	}

	// Total amount = recipients * amount per wallet
	totalAmount := new(big.Float).Mul(amountPerWallet, big.NewFloat(float64(totalRecipients)))

	// Batch count = ceil(recipients / maxBatch)
	batchCount := 0
	if totalRecipients > 0 {
		batchCount = (totalRecipients + maxBatch - 1) / maxBatch
	}

	// Estimated gas per tx
	gasPerTx := baseGasPerTx
	if totalRecipients > 0 {
		recipientsPerBatch := maxBatch
		if totalRecipients < maxBatch {
			recipientsPerBatch = totalRecipients
		}
		gasPerTx = baseGasPerTx + (recipientsPerBatch * defaultGasPerRecipient)
	}

	// Parse gas price from RPC
	gasPriceGweiFloat := parseGasPrice(fc.gasPriceGwei)

	// Gas cost in native token = batchCount * gasPerTx * gasPriceGwei / 10^9
	tenPow9 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil))
	gasCost := new(big.Float).Mul(big.NewFloat(float64(batchCount)), big.NewFloat(float64(gasPerTx)))
	gasCost = new(big.Float).Mul(gasCost, gasPriceGweiFloat)
	gasCost = new(big.Float).Quo(gasCost, tenPow9)

	// Max total funds = totalAmount + gasCost
	maxTotalFunds := new(big.Float).Add(totalAmount, gasCost)

	// Format gas price display
	gasPriceDisplay := "-"
	if fc.gasPriceGwei != "" && fc.gasPriceGwei != "0" {
		gasPriceDisplay = fc.gasPriceGwei + " gwei"
	}

	// Header
	header := fmt.Sprintf(" BULK %s TRANSFER SUMMARY", strings.ToUpper(nativeToken))
	sb.WriteString(SectionHeaderStyle.Width(fc.width).Render(header) + "\n")

	sb.WriteString(fmt.Sprintf("  %-18s: %s addresses\n", "Total Recipients", formatNumber(totalRecipients)))
	sb.WriteString(fmt.Sprintf("  %-18s: %s %s\n", "Amount/Wallet", amountPerWalletStr, nativeToken))
	sb.WriteString(fmt.Sprintf("  %-18s: %s %s\n", fmt.Sprintf("Total %s Amount", nativeToken), formatBigFloat(totalAmount), nativeToken))
	sb.WriteString(fmt.Sprintf("  %-18s: %d transactions\n", "Total Batch Tx", batchCount))
	sb.WriteString(fmt.Sprintf("  %-18s: %d units\n", "Estimated Gas/Tx", gasPerTx))
	sb.WriteString(fmt.Sprintf("  %-18s: %s\n", "Max Gas Fee", gasPriceDisplay))
	sb.WriteString("  " + MutedStyle.Render(strings.Repeat("─", max(0, fc.width-6))) + "\n")
	sb.WriteString(fmt.Sprintf("  %-18s: ~%s %s\n", "Est. Gas Cost", formatBigFloat(gasCost), nativeToken))
	sb.WriteString(fmt.Sprintf("  %-18s: ~%s %s (Amount + Gas)\n", "MAX TOTAL FUNDS", formatBigFloat(maxTotalFunds), nativeToken))
	return sb.String()
}

// parseInt parses a string as int, returning 0 on failure
func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// parseGasPrice parses a gas price string (e.g. "125.686") as a big.Float
func parseGasPrice(s string) *big.Float {
	if s == "" {
		return big.NewFloat(0)
	}
	f, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven)
	if err != nil {
		return big.NewFloat(0)
	}
	return f
}

// formatBigFloat formats a big.Float to a string, trimming trailing zeros
func formatBigFloat(f *big.Float) string {
	s := fmt.Sprintf("%.18f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" {
		return "0"
	}
	return s
}

// formatNumber formats an integer with comma separators
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
