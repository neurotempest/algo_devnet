package test

import (
	"testing"
	"context"
	"os"
	"flag"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/stretchr/testify/require"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/mnemonic"
)

var (
	algodHost = flag.String("algod_host", "http://localhost:4001", "Host of algod client")
	algodTokenPath = flag.String("algod_token_path", "../algorand/algod.token", "Path to algod token")
	kmdHost = flag.String("kmd_host", "http://localhost:4002", "Host of kmd client")
	kmdTokenPath = flag.String("kmd_token_path", "../algorand/kmd.token", "Path to kmd token")
)

func algodClient(t *testing.T) *algod.Client {

	token, err := os.ReadFile(*algodTokenPath)
	require.NoError(t, err)

	c, err := algod.MakeClient(*algodHost, string(token))
	require.NoError(t, err)

	return c
}

func kmdClient(t *testing.T) kmd.Client {

	token, err := os.ReadFile(*kmdTokenPath)
	require.NoError(t, err)

	c, err := kmd.MakeClient(*kmdHost, string(token))
	require.NoError(t, err)

	return c
}

func TestGetPrivateKeyFromKMD(t *testing.T) {

	kmd := kmdClient(t)

	w, err := kmd.ListWallets()
	require.NoError(t, err)
	fmt.Printf("%+v\n", w)

	handle, err := kmd.InitWalletHandle(w.Wallets[0].ID, "")
	require.NoError(t, err)

	k, err := kmd.ListKeys(handle.WalletHandleToken)
	require.NoError(t, err)
	fmt.Printf("%+v\n", k)

	pk, err := kmd.ExportKey(handle.WalletHandleToken, "", k.Addresses[0])
	require.NoError(t, err)
	fmt.Printf("%+v\n", pk)
}

func TestGetBalanceForAddressFromKMD(t *testing.T) {

	kmd := kmdClient(t)
	w, err := kmd.ListWallets()
	require.NoError(t, err)
	require.Greater(t, len(w.Wallets), 0)

	handle, err := kmd.InitWalletHandle(w.Wallets[0].ID, "")
	require.NoError(t, err)

	keys, err := kmd.ListKeys(handle.WalletHandleToken)
	require.NoError(t, err)
	require.Greater(t, len(keys.Addresses), 0)

	algod := algodClient(t)
	info, err := algod.AccountInformation(keys.Addresses[0]).Do(context.Background())
	require.NoError(t, err)

	fmt.Printf("Account info:\n%#v\n", info)
}

func TestSendFromKMDAccountToStandaloneWallet(t *testing.T) {

	account := crypto.GenerateAccount()
	passphrase, err := mnemonic.FromPrivateKey(account.PrivateKey)
	require.NoError(t, err)

	fmt.Println("Created account addr:", account.Address.String(), "\npassphrase:", passphrase)

	kmd := kmdClient(t)
	w, err := kmd.ListWallets()
	require.NoError(t, err)
	require.Greater(t, len(w.Wallets), 0)

	walletPassword := ""
	handleResp, err := kmd.InitWalletHandle(w.Wallets[0].ID, walletPassword)
	require.NoError(t, err)
	handle := handleResp.WalletHandleToken

	keys, err := kmd.ListKeys(handle)
	require.NoError(t, err)
	require.Greater(t, len(keys.Addresses), 0)

	algod := algodClient(t)
	txParams, err := algod.SuggestedParams().Do(context.Background())
	require.NoError(t, err)
	tx, err := future.MakePaymentTxn(
		keys.Addresses[0],
		account.Address.String(),
		1000000,
		[]byte("hello world"),
		"",
		txParams,
	)
	require.NoError(t, err)

	signResp, err := kmd.SignTransaction(handle, walletPassword, tx)
	require.NoError(t, err)

	txID, err := algod.SendRawTransaction(signResp.SignedTransaction).Do(context.Background())
	require.NoError(t, err)

	_, err = future.WaitForConfirmation(algod, txID, 4, context.Background())
	require.NoError(t, err)
}
