package ibc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// This test is meant to be used as a basic interchaintest tutorial.
// Code snippets are broken down in ./docs/upAndRunning.md
func TestBoot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	// Chain Factory
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{Name: "juno", Version: "v11.0.0", ChainConfig: ibc.ChainConfig{
			GasPrices: "0.0ujuno",
		}},
		{Name: "osmosis", Version: "v11.0.0"},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	gaia, osmosis := chains[0], chains[1]

	// Relayer Factory
	client, network := interchaintest.DockerSetup(t)
	r := interchaintest.NewBuiltinRelayerFactory(ibc.Hermes, zaptest.NewLogger(t)).Build(
		t, client, network)

	// Prep Interchain
	const ibcPath = "gaia-osmo-demo"
	ic := interchaintest.NewInterchain().
		AddChain(gaia).
		AddChain(osmosis).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  gaia,
			Chain2:  osmosis,
			Relayer: r,
			Path:    ibcPath,
		})

	// Log location
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false},
	),
	)

	// Create and Fund User Wallets
	fundAmount := int64(10_000_000)
	_, err = interchaintest.GetAndFundTestUserWithMnemonic(ctx, "default", "dilemma imitate split detect useful creek cart sort grow essence fish husband seven hollow envelope wedding host dry permit game april present panic move", fundAmount, gaia)
	require.NoError(t, err)
	_, err = interchaintest.GetAndFundTestUserWithMnemonic(ctx, "default", "settle gas lobster judge silk stem act shoulder pluck waste pistol word comfort require early mouse provide marine butter crowd clock tube move wool", fundAmount, osmosis)
	require.NoError(t, err)

	r.StartRelayer(ctx, eRep)

	time.Sleep(120 * time.Minute)

	r.StopRelayer(ctx, eRep)

}
