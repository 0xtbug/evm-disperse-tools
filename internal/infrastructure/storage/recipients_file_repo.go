package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RecipientsFileRepo implements RecipientRepository using JSON files
type RecipientsFileRepo struct {
	filepath string
}

// NewRecipientsFileRepo creates a new RecipientsFileRepo
func NewRecipientsFileRepo(filepath string) *RecipientsFileRepo {
	return &RecipientsFileRepo{
		filepath: filepath,
	}
}

// Load loads recipients from the JSON file
func (r *RecipientsFileRepo) Load() ([]string, error) {
	data, err := os.ReadFile(r.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read recipients file: %w", err)
	}

	var recipients []string
	if err := json.Unmarshal(data, &recipients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
	}

	return recipients, nil
}

// LoadRecipientsFromTextFile loads recipients from a text file (one address per line)
func LoadRecipientsFromTextFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var recipients []string
	for _, line := range lines {
		addr := strings.TrimSpace(line)
		if addr != "" && !strings.HasPrefix(addr, "#") {
			recipients = append(recipients, addr)
		}
	}

	return recipients, nil
}

type walletManagerEntry struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"` // ignored when loading recipients
}

type walletManagerData struct {
	Wallets []walletManagerEntry `json:"wallets"`
}

// LoadAddressesFromWalletManager loads only addresses from configs/wallets.json
// Private keys are never read into memory for this use case
func LoadAddressesFromWalletManager(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read wallet manager file: %w", err)
	}

	var wf walletManagerData
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse wallet manager file: %w", err)
	}

	var addresses []string
	for _, w := range wf.Wallets {
		if w.Address != "" {
			addresses = append(addresses, w.Address)
		}
	}

	return addresses, nil
}

// WalletListInfo holds metadata about a wallet list file
type WalletListInfo struct {
	Name     string // display name (filename without extension)
	Path     string // full path to the file
	NumAddrs int    // number of addresses in the list
}

// ListWalletFiles scans the wallets directory and returns info about all wallet list files.
func ListWalletFiles(walletsDir string) ([]WalletListInfo, error) {
	var lists []WalletListInfo

	// Ensure the wallets directory exists
	if err := os.MkdirAll(walletsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wallets directory: %w", err)
	}

	entries, err := os.ReadDir(walletsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue
		}
		path := filepath.Join(walletsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var wf walletManagerData
		if json.Unmarshal(data, &wf) != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		lists = append(lists, WalletListInfo{
			Name:     name,
			Path:     path,
			NumAddrs: len(wf.Wallets),
		})
	}

	return lists, nil
}
