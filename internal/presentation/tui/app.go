package tui

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/port"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/storage"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/update"
	"github.com/0xtbug/evm-disperse-tools/internal/version"
)

// DisperseFunc executes a disperse operation and returns the txHash
// DisperseFuncResult holds the outcome of a disperse execution including on-chain receipt
type DisperseFuncResult struct {
	TxHash      string
	BlockNumber uint64
	GasUsed     uint64
	// Batch info
	BatchCount    int
	BatchTxHashes []string
}

// DisperseFunc is the function signature for executing a disperse operation
type DisperseFunc func(ctx context.Context, mode string, chainKey string, recipients []string, amount string, token string) (*DisperseFuncResult, error)

// Page identifiers
type page int

const (
	pageMenu page = iota
	pageDisperseNative
	pageDisperseERC20
	pageWalletManager
	pageSettings
	pageReports
	pageFeeCalculator
)

// Focus pane identifiers
type focusPane int

const (
	focusMain focusPane = iota
	focusActivity
	focusStats
	focusCount
)

// AppModel is the main application model with full layout
type AppModel struct {
	// Layout
	width  int
	height int
	page   page

	// Panes
	menu      MenuModel
	logs      []string
	focusPane focusPane

	// State
	mouseEnabled bool
	copyFeedback string

	// Forms
	disperseNativeForm *DisperseFormScreen
	disperseERC20Form  *DisperseFormScreen
	walletManager      *WalletManagerModel
	settingsScreen     *SettingsScreen
	reportsScreen      *ReportsScreen
	feeCalculator      *FeeCalculatorScreen

	// Chain configuration (loaded from YAML files)
	chains []*config.ChainConfig

	// App configuration
	appCfg *config.AppConfig

	// Disperse execution
	disperseFunc DisperseFunc

	// Cached RPC client for health monitoring
	monitorClient *ethclient.Client
	monitorRPCURL string

	// Realtime RPC State
	rpcStatus     bool
	blockHeight   uint64
	senderBalance string
	senderAddr    string
	pingLatency   string
	gasPrice      string

	// Activity log scrolling
	activityScrollOffset int

	// Update check result
	updateResult *update.Result
}

// NewAppModel creates a new application model
func NewAppModel(chains []*config.ChainConfig, appCfg *config.AppConfig, disperseFunc DisperseFunc, senderAddr string, reportRepo port.ReportRepository) *AppModel {
	chainNames := make([]string, len(chains))
	for i, c := range chains {
		chainNames[i] = c.Name
	}

	// Resolve default chain key to chain name for disperse forms
	defaultChainName := ""
	for _, c := range chains {
		if c.Key == appCfg.App.DefaultChain {
			defaultChainName = c.Name
			break
		}
	}

	// Load wallet lists
	walletLists, _ := storage.ListWalletFiles(filepath.Join("configs", "wallets"))

	am := &AppModel{
		page:               pageMenu,
		menu:               NewMenuModel(),
		mouseEnabled:       true,
		focusPane:          focusMain,
		disperseNativeForm: NewDisperseFormScreen("Native", chainNames, walletLists, appCfg.App.DefaultAmount, defaultChainName),
		disperseERC20Form:  NewDisperseFormScreen("ERC20", chainNames, walletLists, appCfg.App.DefaultAmount, defaultChainName),
		walletManager:      NewWalletManagerModel(),
		settingsScreen:     NewSettingsScreen(appCfg, chains),
		reportsScreen:      NewReportsScreen(reportRepo),
		feeCalculator:      NewFeeCalculatorScreen(chains, appCfg, defaultChainName),
		chains:             chains,
		appCfg:             appCfg,
		disperseFunc:       disperseFunc,
		senderAddr:         senderAddr,
	}
	am.appendLog("EVM Disperse TUI Ready")
	return am
}

// timestamp returns a styled timestamp string
func timestamp() string {
	return LogTimestampStyle.Render(time.Now().Format("15:04:05"))
}

// rpcStatusMsg is sent when an RPC health check completes
type rpcStatusMsg struct {
	ok       bool
	block    uint64
	latency  string
	balance  string
	gasPrice string
}

// rpcTickMsg triggers a periodic RPC health check
type rpcTickMsg struct{}

// doRPCHealthCheck performs an RPC connectivity check using a cached client
func doRPCHealthCheck(client *ethclient.Client, senderAddr string) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return rpcStatusMsg{ok: false}
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		block, err := client.BlockNumber(ctx)
		if err != nil {
			return rpcStatusMsg{ok: false}
		}

		latency := fmt.Sprintf("%dms", time.Since(start).Milliseconds())

		// Fetch gas price
		var gasPriceStr string
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err == nil {
			gasFloat := new(big.Float).Quo(
				new(big.Float).SetInt(gasPrice),
				new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
			)
			gasPriceStr = fmt.Sprintf("%.3f Gwei", gasFloat)
		}

		var balanceStr string
		if senderAddr != "" {
			bal, err := client.BalanceAt(ctx, common.HexToAddress(senderAddr), nil)
			if err == nil {
				balFloat := new(big.Float).Quo(
					new(big.Float).SetInt(bal),
					new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
				)
				balanceStr = fmt.Sprintf("%.6f", balFloat)
			}
		}

		return rpcStatusMsg{ok: true, block: block, latency: latency, balance: balanceStr, gasPrice: gasPriceStr}
	}
}

// scheduleRPCTick schedules the next RPC health check in 5 seconds
func scheduleRPCTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return rpcTickMsg{}
	})
}

// getMonitorRPCURL returns the RPC URL to monitor (default chain or first chain)
func (am *AppModel) getMonitorRPCURL() string {
	if c := am.getMonitorChain(); c != nil {
		return c.RPCURL
	}
	return ""
}

// getMonitorChain returns the chain config for the default chain (or first chain).
// This is used for the stats sidebar to always reflect the monitored chain.
func (am *AppModel) getMonitorChain() *config.ChainConfig {
	if am.appCfg != nil && am.appCfg.App.DefaultChain != "" {
		for _, c := range am.chains {
			if c.Key == am.appCfg.App.DefaultChain {
				return c
			}
		}
	}
	if len(am.chains) > 0 {
		return am.chains[0]
	}
	return nil
}

// getOrCreateMonitorClient returns a cached ethclient for the current monitor chain,
// creating a new connection only when the RPC URL changes.
func (am *AppModel) getOrCreateMonitorClient() *ethclient.Client {
	rpcURL := am.getMonitorRPCURL()
	if rpcURL == "" {
		return nil
	}
	if am.monitorClient != nil && am.monitorRPCURL == rpcURL {
		return am.monitorClient
	}
	// Close old client if URL changed
	if am.monitorClient != nil {
		am.monitorClient.Close()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil
	}
	am.monitorClient = client
	am.monitorRPCURL = rpcURL
	return client
}

// Init initializes the model
func (am *AppModel) Init() tea.Cmd {
	return tea.Batch(tea.EnableMouseCellMotion, doRPCHealthCheck(am.getOrCreateMonitorClient(), am.senderAddr), checkForUpdateCmd())
}

// checkForUpdateCmd returns a command that checks for application updates.
func checkForUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		result, err := update.Check()
		if err != nil {
			return updateResultMsg{}
		}
		return updateResultMsg{
			latestVersion: result.LatestVersion,
			hasUpdate:     result.HasUpdate,
			releaseURL:    result.ReleaseURL,
		}
	}
}

// deriveSenderAddress re-derives the sender address from the current private key
// based on the key mode (global or per_chain) and the default chain.
// This should be called whenever the private key changes.
func (am *AppModel) deriveSenderAddress() {
	privKey := am.appCfg.GetPrivateKeyForChain(am.appCfg.App.DefaultChain)
	if privKey == "" && len(am.chains) > 0 {
		privKey = am.appCfg.GetPrivateKeyForChain(am.chains[0].Key)
	}
	if privKey == "" {
		am.senderAddr = ""
		return
	}
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privKey, "0x"))
	if err != nil {
		am.senderAddr = ""
		return
	}
	am.senderAddr = crypto.PubkeyToAddress(*pk.Public().(*ecdsa.PublicKey)).Hex()
}

// fetchGasPriceForFeeCalc returns a command that fetches the current gas price
// from the selected chain's RPC for the fee calculator.
func (am *AppModel) fetchGasPriceForFeeCalc() tea.Cmd {
	if am.feeCalculator == nil {
		return nil
	}
	chain := am.feeCalculator.getCurrentChain()
	if chain == nil {
		return nil
	}

	rpcURL := chain.RPCURL
	chainKey := chain.Key

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := ethclient.DialContext(ctx, rpcURL)
		if err != nil {
			return feeCalcGasPriceMsg{chainKey: chainKey, gasPrice: ""}
		}
		defer client.Close()

		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return feeCalcGasPriceMsg{chainKey: chainKey, gasPrice: ""}
		}

		gasFloat := new(big.Float).Quo(
			new(big.Float).SetInt(gasPrice),
			new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)),
		)
		gasPriceStr := fmt.Sprintf("%.3f", gasFloat)

		return feeCalcGasPriceMsg{chainKey: chainKey, gasPrice: gasPriceStr}
	}
}

// Update handles messages
func (am *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		am.width = msg.Width
		am.height = msg.Height
		return am, nil

	case tea.KeyMsg:
		key := msg.String()

		// Global quit
		if key == "ctrl+c" {
			return am, tea.Quit
		}

		// Settings page handles its own keys (esc exits input mode first, then goes back)
		if am.page == pageSettings && am.focusPane == focusMain {
			newSS, cmd := am.settingsScreen.Update(msg)
			am.settingsScreen = newSS
			return am, cmd
		}

		// Esc to go back to menu from other pages
		if key == "esc" {
			if am.page != pageMenu {
				am.page = pageMenu
				am.focusPane = focusMain
				return am, nil
			}
		}

		// Activity log scrolling and clearing (when activity pane is focused)
		if am.focusPane == focusActivity {
			switch {
			case key == "up" || msg.Type == tea.KeyUp:
				am.scrollActivityUp()
				return am, nil
			case key == "down" || msg.Type == tea.KeyDown:
				am.scrollActivityDown()
				return am, nil
			case key == "c":
				am.logs = nil
				am.activityScrollOffset = 0
				am.appendLog("Activity log cleared")
				return am, nil
			}
		}

		// Reports page scrolling (when main pane is focused)
		if am.page == pageReports && am.focusPane == focusMain {
			switch {
			case key == "up" || msg.Type == tea.KeyUp:
				am.reportsScreen.ScrollUp()
				return am, nil
			case key == "down" || msg.Type == tea.KeyDown:
				am.reportsScreen.ScrollDown()
				return am, nil
			case key == "r":
				am.reportsScreen.LoadReports()
				am.appendLog("Reports refreshed — " + am.reportsScreen.Summary())
				return am, nil
			}
		}

		// Menu page shortcuts
		if am.page == pageMenu {
			if key == "q" {
				return am, tea.Quit
			}
			// Clear activity logs
			if key == "c" {
				am.logs = nil
				am.activityScrollOffset = 0
				am.appendLog("Activity log cleared")
				return am, nil
			}
			// Wallet manager
			if key == "w" {
				am.page = pageWalletManager
				return am, nil
			}
			// Settings
			if key == "s" {
				am.page = pageSettings
				return am, nil
			}
			// Toggle mouse mode
			if key == "m" {
				am.mouseEnabled = !am.mouseEnabled
				if am.mouseEnabled {
					return am, tea.EnableMouseCellMotion
				}
				return am, tea.DisableMouse
			}
			// Switch pane focus
			if key == "tab" {
				am.focusPane = (am.focusPane + 1) % focusCount
				return am, nil
			}
		}

		// Handle menu navigation when on menu page
		if am.page == pageMenu && am.focusPane == focusMain {
			newMenu, cmd := am.menu.Update(msg)
			am.menu = newMenu
			if cmd != nil {
				return am, am.interceptMenuCmd(cmd)
			}
		}

		// Form page navigation (only when main pane is focused)
		if (am.page == pageDisperseNative || am.page == pageDisperseERC20) && am.focusPane == focusMain {
			var form *DisperseFormScreen
			if am.page == pageDisperseNative {
				form = am.disperseNativeForm
			} else {
				form = am.disperseERC20Form
			}
			if form.Update(key) {
				return am, nil
			}

			// Ctrl+D to execute disperse
			if key == "ctrl+d" {
				return am.handleDisperseExecute()
			}
		}

		// Fee calculator page navigation (only when main pane is focused)
		if am.page == pageFeeCalculator && am.focusPane == focusMain {
			oldChainIdx := am.feeCalculator.GetSelectedChainIdx()
			if am.feeCalculator.Update(key) {
				if am.feeCalculator.GetSelectedChainIdx() != oldChainIdx {
					return am, am.fetchGasPriceForFeeCalc()
				}
				return am, nil
			}
		}

		// Wallet manager page navigation (only when main pane is focused)
		if am.page == pageWalletManager && am.focusPane == focusMain {
			newWM, cmd := am.walletManager.Update(msg)
			am.walletManager = newWM
			return am, cmd
		}

	case tea.MouseMsg:
		if !am.mouseEnabled {
			break
		}
		if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
			break
		}

		// Handle pane switching via clicks
		headerStartY, mainStartY, activityStartY, _ := am.calculateLayoutPositions()

		// Check if click is in header/stats pane area
		if msg.Y >= headerStartY && msg.Y < mainStartY {
			am.focusPane = focusStats
		} else if msg.Y >= mainStartY && msg.Y < activityStartY {
			am.focusPane = focusMain

			// Handle menu item clicks
			if am.page == pageMenu {
				menuStartY := mainStartY + 2 // border + padding
				clicked, cmd := am.menu.HandleMouseClick(int(msg.Y), menuStartY)
				if clicked && cmd != nil {
					return am, am.interceptMenuCmd(cmd)
				}
			}
		} else if msg.Y >= activityStartY {
			am.focusPane = focusActivity
		}

	case runToolMsg:
		return am.handleToolRun(msg)

	case toolDoneMsg:
		am.appendLog(fmt.Sprintf("Task finished: %s", msg.output))
		return am, nil

	case logMsg:
		am.appendLog(msg.message)
		return am, nil

	case logAppendMsg:
		am.appendLog(string(msg))
		return am, nil

	case copyFeedbackMsg:
		if am.copyFeedback == string(msg) {
			am.copyFeedback = ""
		}
		return am, nil

	case walletGeneratedMsg:
		if am.page == pageWalletManager {
			am.appendLog(fmt.Sprintf("Generated %d wallet(s)", len(msg.wallets)))
			newWM, cmd := am.walletManager.Update(msg)
			am.walletManager = newWM
			return am, cmd
		}

	case walletSavedMsg:
		if am.page == pageWalletManager {
			if strings.HasPrefix(msg.path, "error") {
				am.appendReport(fmt.Sprintf("Wallet save failed: %s", msg.path), "error")
			} else {
				am.appendReport(fmt.Sprintf("Wallets saved to %s", msg.path), "success")
				// Refresh wallet lists in disperse forms so new wallets appear immediately
				am.refreshWalletLists()
			}
			newWM, cmd := am.walletManager.Update(msg)
			am.walletManager = newWM
			return am, cmd
		}

	case goToMenuMsg:
		am.page = pageMenu
		am.focusPane = focusMain
		return am, nil

	case privateKeyChangedMsg:
		// Re-derive sender address from the updated private key
		am.deriveSenderAddress()
		am.appendLog("Private key changed — updated sender address")
		if am.senderAddr != "" {
			am.appendLog("Sender: " + am.senderAddr)
		}
		return am, doRPCHealthCheck(am.getOrCreateMonitorClient(), am.senderAddr)

	case settingsSavedMsg2:
		am.appendLog("Settings saved to configs/app.yaml")
		// Re-derive sender address from the (possibly updated) private key
		am.deriveSenderAddress()
		// Update disperse forms with the new default chain
		for _, c := range am.chains {
			if c.Key == am.appCfg.App.DefaultChain {
				am.disperseNativeForm.UpdateDefaultChain(c.Name)
				am.disperseERC20Form.UpdateDefaultChain(c.Name)
				break
			}
		}
		// Update disperse forms with the new default amount
		am.disperseNativeForm.UpdateDefaultAmount(am.appCfg.App.DefaultAmount)
		am.disperseERC20Form.UpdateDefaultAmount(am.appCfg.App.DefaultAmount)
		// Trigger immediate RPC health check with the (possibly updated) sender address
		return am, doRPCHealthCheck(am.getOrCreateMonitorClient(), am.senderAddr)

	case settingsErrMsg:
		am.appendLog(fmt.Sprintf("Settings save failed: %s", string(msg)))
		return am, nil

	case feeCalcGasPriceMsg:
		if am.feeCalculator != nil {
			if chain := am.feeCalculator.getCurrentChain(); chain != nil && chain.Key == msg.chainKey {
				am.feeCalculator.SetGasPrice(msg.gasPrice)
			}
		}
		return am, nil

	case disperseResultMsg:
		if msg.err != nil {
			if msg.batchCount > 1 {
				// Batch operation with partial failure
				am.appendReport(fmt.Sprintf("Disperse %s batch FAILED on %s — completed %d/%d batches — %s",
					msg.mode, msg.chainName, len(msg.batchTxHashes), msg.batchCount, msg.err.Error()), "error")
				for i, txHash := range msg.batchTxHashes {
					am.appendReport(fmt.Sprintf("  Batch %d/%d tx: %s", i+1, msg.batchCount, txHash), "info")
				}
			} else if msg.blockNumber > 0 {
				am.appendReport(fmt.Sprintf("Disperse %s REVERTED on %s (block %d, gas %d) — tx: %s — %s",
					msg.mode, msg.chainName, msg.blockNumber, msg.gasUsed, msg.txHash, msg.err.Error()), "error")
			} else {
				am.appendReport(fmt.Sprintf("Disperse %s failed: %s", msg.mode, msg.err.Error()), "error")
			}
		} else if msg.batchCount > 1 {
			// Multi-batch success
			am.appendReport(fmt.Sprintf("Disperse %s ALL %d BATCHES CONFIRMED on %s — gas %d",
				msg.mode, msg.batchCount, msg.chainName, msg.gasUsed), "success")
			for i, txHash := range msg.batchTxHashes {
				am.appendReport(fmt.Sprintf("  Batch %d/%d tx: %s", i+1, msg.batchCount, txHash), "info")
			}
		} else {
			am.appendReport(fmt.Sprintf("Disperse %s CONFIRMED on %s (block %d, gas %d) — tx: %s",
				msg.mode, msg.chainName, msg.blockNumber, msg.gasUsed, msg.txHash), "success")
		}
		return am, nil

	case rpcTickMsg:
		if am.page == pageFeeCalculator {
			return am, tea.Batch(
				doRPCHealthCheck(am.getOrCreateMonitorClient(), am.senderAddr),
				am.fetchGasPriceForFeeCalc(),
			)
		}
		return am, doRPCHealthCheck(am.getOrCreateMonitorClient(), am.senderAddr)

	case rpcStatusMsg:
		am.rpcStatus = msg.ok
		if msg.ok {
			am.blockHeight = msg.block
			am.pingLatency = msg.latency
			am.senderBalance = msg.balance
			am.gasPrice = msg.gasPrice
		} else {
			am.blockHeight = 0
			am.pingLatency = ""
			am.senderBalance = ""
			am.gasPrice = ""
		}
		return am, scheduleRPCTick()
	}

	return am, nil
}

// handleToolRun processes menu selection
func (am *AppModel) handleToolRun(msg runToolMsg) (tea.Model, tea.Cmd) {
	switch msg.toolID {
	case 1:
		am.page = pageDisperseNative
		am.refreshWalletLists()
	case 2:
		am.page = pageDisperseERC20
		am.refreshWalletLists()
	case 3:
		am.page = pageReports
		am.reportsScreen.LoadReports()
		am.appendLog(am.reportsScreen.Summary())
	case 4:
		am.page = pageFeeCalculator
		return am, am.fetchGasPriceForFeeCalc()
	}
	return am, nil
}

// handleDisperseExecute handles the Ctrl+D execution of a disperse operation
func (am *AppModel) handleDisperseExecute() (tea.Model, tea.Cmd) {
	var form *DisperseFormScreen
	var mode string
	if am.page == pageDisperseNative {
		form = am.disperseNativeForm
		mode = "Native"
	} else {
		form = am.disperseERC20Form
		mode = "ERC20"
	}

	if !form.IsValid() {
		missing := "check chain, wallet list, amount"
		if mode == "ERC20" {
			missing = "check chain, wallet list, token, amount"
		}
		am.appendLog("Disperse validation failed — " + missing)
		return am, nil
	}

	chain := am.getCurrentChain()
	recipients := form.GetRecipients()
	amount := form.GetAmount()
	token := form.GetToken()

	chainName := "unknown"
	chainKey := ""
	if chain != nil {
		chainName = chain.Name
		chainKey = chain.Key
	}

	am.appendReport(fmt.Sprintf("Disperse %s executing — %s to %d recipients on %s",
		mode, amount, len(recipients), chainName), "info")

	// Capture values for the async command
	execMode := mode
	execChainKey := chainKey
	execRecipients := make([]string, len(recipients))
	copy(execRecipients, recipients)
	execAmount := amount
	execToken := token
	execChainName := chainName
	execFunc := am.disperseFunc

	return am, func() tea.Msg {
		if execFunc == nil {
			return disperseResultMsg{
				err:       fmt.Errorf("disperse not configured — restart the app"),
				mode:      execMode,
				chainName: execChainName,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		result, err := execFunc(ctx, execMode, execChainKey, execRecipients, execAmount, execToken)
		msg := disperseResultMsg{
			err:       err,
			mode:      execMode,
			chainName: execChainName,
		}
		if result != nil {
			msg.txHash = result.TxHash
			msg.blockNumber = result.BlockNumber
			msg.gasUsed = result.GasUsed
			msg.batchCount = result.BatchCount
			msg.batchTxHashes = result.BatchTxHashes
		}
		return msg
	}
}

// interceptMenuCmd wraps menu commands to add tool name
func (am *AppModel) interceptMenuCmd(cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		msg := cmd()
		if rt, ok := msg.(runToolMsg); ok {
			for _, e := range menuEntries {
				if e.id == rt.toolID {
					rt.toolName = e.label
					break
				}
			}
			return rt
		}
		return msg
	}
}

// appendLog adds a log entry with timestamp
func (am *AppModel) appendLog(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	formatted := fmt.Sprintf("%s %s", timestamp(), msg)
	am.logs = append(am.logs, formatted)
	if len(am.logs) > 200 {
		am.logs = am.logs[1:]
	}
	am.activityScrollOffset = 0
}

// appendReport adds a log entry to the activity pane with a specific level.
// Reports are persisted to disk by the disperse use case and loaded separately.
func (am *AppModel) appendReport(msg string, level string) {
	am.appendLog(msg)
}

// AddLog adds a log entry (public)
func (am *AppModel) AddLog(msg string) {
	am.appendLog(msg)
}

// refreshWalletLists reloads wallet list files from disk and updates both disperse forms.
func (am *AppModel) refreshWalletLists() {
	walletLists, err := storage.ListWalletFiles(filepath.Join("configs", "wallets"))
	if err != nil {
		am.appendLog(fmt.Sprintf("Failed to refresh wallet lists: %v", err))
		return
	}
	am.disperseNativeForm.UpdateWalletLists(walletLists)
	am.disperseERC20Form.UpdateWalletLists(walletLists)
}

// getCurrentChain returns the currently selected chain config
func (am *AppModel) getCurrentChain() *config.ChainConfig {
	var selectedIdx int
	switch am.page {
	case pageDisperseNative:
		selectedIdx = am.disperseNativeForm.GetSelectedChainIdx()
	case pageDisperseERC20:
		selectedIdx = am.disperseERC20Form.GetSelectedChainIdx()
	default:
		selectedIdx = 0
	}

	if len(am.chains) == 0 || selectedIdx >= len(am.chains) {
		return nil
	}
	return am.chains[selectedIdx]
}

// scrollActivityUp scrolls up to older logs
func (am *AppModel) scrollActivityUp() {
	am.activityScrollOffset++
	maxScroll := max(0, len(am.logs)-1)
	if am.activityScrollOffset > maxScroll {
		am.activityScrollOffset = maxScroll
	}
}

// scrollActivityDown scrolls down to newer logs
func (am *AppModel) scrollActivityDown() {
	if am.activityScrollOffset > 0 {
		am.activityScrollOffset--
	}
}

// getVisibleLogs returns the log entries that fit in the visible area
func (am *AppModel) getVisibleLogs(maxLines int) []string {
	if len(am.logs) == 0 {
		return nil
	}

	// Clamp scroll offset
	maxOffset := max(0, len(am.logs)-maxLines)
	if am.activityScrollOffset > maxOffset {
		am.activityScrollOffset = maxOffset
	}

	endIdx := len(am.logs) - am.activityScrollOffset
	startIdx := endIdx - maxLines
	if startIdx < 0 {
		startIdx = 0
	}

	return am.logs[startIdx:endIdx]
}

// View renders the complete application layout
func (am *AppModel) View() string {
	if am.width == 0 || am.height == 0 {
		return "Loading..."
	}

	headerH, mainH, logH, _ := am.calculateLayout()

	// Header with ASCII art and stats
	topInnerWidth, topInnerHeight := max(0, am.width-4), max(0, headerH-2)
	leftWidth := topInnerWidth / 2
	rightWidth := topInnerWidth - leftWidth

	asciiArt := SidebarTitleStyle.Render(am.headerInfo())
	titleBox := lipgloss.Place(leftWidth, topInnerHeight, lipgloss.Center, lipgloss.Center, asciiArt)
	statsBox := lipgloss.Place(rightWidth, topInnerHeight, lipgloss.Left, lipgloss.Center, am.statsView(rightWidth))

	topRow := am.paneStyle(focusStats).Width(topInnerWidth).Height(topInnerHeight).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, titleBox, statsBox),
	)

	// Main content pane
	mainInnerWidth, mainInnerHeight := max(0, am.width-4), max(0, mainH-2)
	var content string

	switch am.page {
	case pageMenu:
		content = am.menu.View(mainInnerWidth)
	case pageDisperseNative:
		content = am.disperseNativeForm.View(mainInnerWidth)
	case pageDisperseERC20:
		content = am.disperseERC20Form.View(mainInnerWidth)
	case pageWalletManager:
		content = am.walletManager.View(mainInnerWidth)
	case pageSettings:
		content = am.settingsScreen.View(mainInnerWidth)
	case pageReports:
		content = am.reportsScreen.View(mainInnerWidth)
	case pageFeeCalculator:
		content = am.feeCalculator.View(mainInnerWidth)
	}

	contentPane := am.paneStyle(focusMain).Width(mainInnerWidth).Height(mainInnerHeight).Render(content)

	// Activity/Log pane
	activityInnerWidth, logInnerHeight := max(0, am.width-4), max(0, logH-2)

	// Show only the logs that fit in the visible area
	maxLogLines := max(1, logInnerHeight-2)
	visibleLogs := am.getVisibleLogs(maxLogLines)

	// Truncate log lines so they never wrap and shift the layout
	maxLogWidth := max(1, activityInnerWidth-4)
	for i, line := range visibleLogs {
		visibleLogs[i] = truncateString(line, maxLogWidth)
	}

	logContent := ""
	if len(am.logs) == 0 {
		logContent = "  No activity yet..."
	} else {
		logContent = strings.Join(visibleLogs, "\n")
	}

	// Show scroll indicator when scrolled up
	activityHeader := " ACTIVITY"
	if am.activityScrollOffset > 0 {
		activityHeader = fmt.Sprintf(" ACTIVITY  ↑%d", am.activityScrollOffset)
	}

	activityPane := am.paneStyle(focusActivity).Width(activityInnerWidth).Height(logInnerHeight).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Width(activityInnerWidth).Render(activityHeader),
			logContent,
		),
	)

	// Combine all panes
	output := lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		contentPane,
		activityPane,
		am.footerView(),
	)

	return output
}

// statsView renders the stats sidebar
func (am *AppModel) statsView(width int) string {
	chain := am.getMonitorChain()
	if chain == nil {
		return MutedStyle.Render("  No chains configured")
	}

	colWidth := width / 2
	maxValWidth := max(10, colWidth-14)

	shortRPC := chain.RPCURL
	if len(shortRPC) > maxValWidth {
		shortRPC = shortRPC[:maxValWidth-3] + "..."
	}

	// Real-time RPC status
	statusColor := lipgloss.Color("#e06c75") // Red for offline
	statusText := "Offline"
	blockStr := "-"
	balanceStr := "-"
	nativeSymbol := chain.GetNativeToken()
	if am.rpcStatus {
		statusColor = lipgloss.Color("#98c379") // Green for online
		statusText = fmt.Sprintf("Online (%s)", am.pingLatency)
		blockStr = fmt.Sprintf("%d", am.blockHeight)
		if am.senderBalance != "" {
			balanceStr = am.senderBalance + " " + nativeSymbol
		}
	}
	rpcStatusRender := lipgloss.NewStyle().Foreground(statusColor).Render(statusText)

	left1 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("RPC"), MetaValStyle.Render(shortRPC))
	left2 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("Chain"), MetaValStyle.Render(chain.Name))
	left3 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("Status"), rpcStatusRender)

	right1 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("Balance"), MetaValStyle.Render(balanceStr))
	right2 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("Block"), MetaValStyle.Render(blockStr))
	gasStr := "-"
	if am.gasPrice != "" {
		gasStr = am.gasPrice
	}
	right3 := fmt.Sprintf(" %s │ %s", SidebarLabelStyle.Render("Gas"), MetaValStyle.Render(gasStr))

	colStyle := lipgloss.NewStyle().Width(colWidth)
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(left1), colStyle.Render(right1)),
		lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(left2), colStyle.Render(right2)),
		lipgloss.JoinHorizontal(lipgloss.Top, colStyle.Render(left3), colStyle.Render(right3)),
	)
}

// footerView renders the status bar with keyboard shortcuts
func (am *AppModel) footerView() string {
	var help string
	switch am.page {
	case pageDisperseNative, pageDisperseERC20:
		help = fmt.Sprintf(" %s back to menu  %s next field  %s prev field  %s change chain  %s execute",
			StatusKeyStyle.Render("Esc"),
			StatusKeyStyle.Render("Tab"),
			StatusKeyStyle.Render("Shift+Tab"),
			StatusKeyStyle.Render("↑↓"),
			StatusKeyStyle.Render("Ctrl+D"))
	case pageWalletManager:
		help = fmt.Sprintf(" %s back  %s generate  %s save JSON  %s export privkeys  %s export addresses",
			StatusKeyStyle.Render("Esc"),
			StatusKeyStyle.Render("G"),
			StatusKeyStyle.Render("S"),
			StatusKeyStyle.Render("P"),
			StatusKeyStyle.Render("A"))
	case pageSettings:
		help = fmt.Sprintf(" %s navigate  %s toggle/edit  %s save  %s back",
			StatusKeyStyle.Render("↑↓"),
			StatusKeyStyle.Render("Enter"),
			StatusKeyStyle.Render("Ctrl+S"),
			StatusKeyStyle.Render("Esc"))
	case pageReports:
		help = fmt.Sprintf(" %s back  %s scroll  %s refresh reports",
			StatusKeyStyle.Render("Esc"),
			StatusKeyStyle.Render("↑↓"),
			StatusKeyStyle.Render("R"))
	case pageFeeCalculator:
		help = fmt.Sprintf(" %s navigate  %s edit  %s back",
			StatusKeyStyle.Render("↑↓/Tab"),
			StatusKeyStyle.Render("Enter"),
			StatusKeyStyle.Render("Esc"))
	default:
		help = fmt.Sprintf(" %s wallet  %s settings  %s mode  %s pane  %s select  %s clear logs  %s quit",
			StatusKeyStyle.Render("w"),
			StatusKeyStyle.Render("s"),
			StatusKeyStyle.Render("m"),
			StatusKeyStyle.Render("Tab"),
			StatusKeyStyle.Render("Enter"),
			StatusKeyStyle.Render("c"),
			StatusKeyStyle.Render("q"))
	}

	modeStr := ""
	if !am.mouseEnabled {
		modeStr = ErrorStyle.Bold(true).Render(" [SELECTION MODE] ")
	} else {
		modeStr = SuccessStyle.Bold(true).Render(" [INTERACTIVE] ")
	}

	if am.copyFeedback != "" {
		modeStr = SuccessStyle.Bold(true).Render(" [COPIED] ") + am.copyFeedback
	}

	return StatusBarStyle.Width(max(0, am.width-3)).Render(modeStr + help)
}

// paneStyle returns the appropriate pane border style based on focus
func (am *AppModel) paneStyle(paneID focusPane) lipgloss.Style {
	if am.focusPane == paneID {
		return ActivePaneStyle.Copy()
	}
	return BasePaneStyle.Copy()
}

// calculateLayout calculates the height of each section
func (am *AppModel) calculateLayout() (headerH, mainH, logH, footerH int) {
	headerH, footerH = 8, 1
	logH = 15
	mainH = am.height - headerH - footerH - logH
	if mainH < 14 {
		mainH = 14
		logH = max(4, am.height-headerH-footerH-mainH)
		mainH = am.height - headerH - footerH - logH
	}
	return
}

// calculateLayoutPositions calculates the Y positions of each section
func (am *AppModel) calculateLayoutPositions() (headerY, mainY, activityY, footerY int) {
	headerY = 0
	headerH, mainH, logH, footerH := am.calculateLayout()
	mainY = headerY + headerH
	activityY = mainY + mainH
	footerY = activityY + logH
	if footerY > am.height-footerH {
		footerY = am.height - footerH
	}
	return
}

// getActivityPaneY returns the Y position where the activity pane starts
func (am *AppModel) getActivityPaneY() int {
	_, _, activityY, _ := am.calculateLayoutPositions()
	return activityY
}

// Placeholder views for other pages
func (am *AppModel) disperseNativeView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Disperse Native Tokens"),
		"",
		"  Select chain, load recipients, and execute.",
		"",
		MutedStyle.Italic(true).Render("Press Esc to return to menu"),
	)
}

func (am *AppModel) disperseERC20View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Disperse ERC20 Tokens"),
		"",
		"  Select token, chain, load recipients, and execute.",
		"",
		MutedStyle.Italic(true).Render("Press Esc to return to menu"),
	)
}

func (am *AppModel) reportsView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("Execution Reports"),
		"",
		"  View past execution history and reports.",
		"",
		MutedStyle.Italic(true).Render("Press Esc to return to menu"),
	)
}

// truncateString truncates a string to maxLen, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return "..." + s[len(s)-(maxLen-3):]
}

// copyFeedbackMsg is used to clear copy feedback after timeout

// headerInfo builds the full header text with ASCII art, version, and update notice.
func (am *AppModel) headerInfo() string {
	info := disperseAscii + "\n"
	info += "Version: " + version.Version + "\n"
	if am.updateResult != nil && am.updateResult.HasUpdate {
		info += "Update:  " + am.updateResult.LatestVersion + " available\n"
	}
	info += "Github: https://github.com/0xtbug/evm-disperse-tools"
	return info
}

type copyFeedbackMsg string

// disperseAscii is the ASCII art header
const disperseAscii = "\n" +
	"┏━╸╻ ╻┏┳┓   ╺┳┓╻┏━┓┏━┓┏━╸┏━┓┏━┓┏━╸\n" +
	"┣╸ ┃┏┛┃┃┃╺━╸ ┃┃┃┗━┓┣━┛┣╸ ┣┳┛┗━┓┣╸\n" +
	"┗━╸┗┛ ╹ ╹   ╺┻┛╹┗━┛╹  ┗━╸╹┗╸┗━┛┗━╸"
