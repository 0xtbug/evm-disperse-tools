package tui

// Navigation messages
type goToMenuMsg struct{}
type goToSettingsMsg struct{}
type goToReportsMsg struct{}

// Action messages
type runToolMsg struct {
	toolID   int
	toolName string
}

type toolDoneMsg struct {
	output string
}

// Log messages
type logMsg struct {
	message string
}

type logAppendMsg string

// Mouse toggle
type toggleMouseMsg struct{}

// Private key changed (emitted when a key field is applied in settings)
type privateKeyChangedMsg struct{}

// Fee calculator gas price result
type feeCalcGasPriceMsg struct {
	chainKey string
	gasPrice string // in gwei, e.g. "125.686"
}

// Update check result
type updateResultMsg struct {
	latestVersion string
	hasUpdate     bool
	releaseURL    string
}

// Disperse execution result
type disperseResultMsg struct {
	txHash      string
	blockNumber uint64
	gasUsed     uint64
	err         error
	mode        string
	chainName   string
	// Batch info
	batchCount    int
	batchIdx      int // 1-based current batch index (for progress)
	batchTxHashes []string
}
