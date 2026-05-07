package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/evm"
)

func TestPolygonRPCConnectivity(t *testing.T) {
	if os.Getenv("POLYGON_RPC_URL") == "" {
		t.Skip("POLYGON_RPC_URL not set")
	}

	rpcURL := os.Getenv("POLYGON_RPC_URL")
	client, err := evm.NewRPCClient(rpcURL)
	if err != nil {
		t.Fatalf("failed to create RPC client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatalf("failed to get chain ID: %v", err)
	}

	if chainID.Uint64() != 137 {
		t.Fatalf("expected chain ID 137, got %d", chainID.Uint64())
	}
}

func TestPolygonConfigValid(t *testing.T) {
	cfg, err := config.LoadChainConfig("../../configs/chains/polygon.yaml")
	if err != nil {
		t.Fatalf("failed to load polygon config: %v", err)
	}

	if cfg.Key != "polygon" {
		t.Fatalf("expected key 'polygon', got %s", cfg.Key)
	}
	if cfg.ChainID != 137 {
		t.Fatalf("expected chain ID 137, got %d", cfg.ChainID)
	}
}
