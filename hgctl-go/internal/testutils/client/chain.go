package client

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ITaskMailbox"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type ChainCaller struct {
	chainId     *big.Int
	taskMailbox *ITaskMailbox.ITaskMailbox
	ethclient   *ethclient.Client
	logger      logger.Logger
	signer      *signer.PrivateKeySigner
}

func NewChainCaller(
	ethclient *ethclient.Client,
	signer *signer.PrivateKeySigner,
	chainConfig config.ChainConfig,
	logger logger.Logger,
) (*ChainCaller, error) {
	logger.Sugar().Debugw("Creating contract caller")

	chainId, err := ethclient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	taskMailbox, err := ITaskMailbox.NewITaskMailbox(common.HexToAddress(chainConfig.TaskMailboxAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskMailbox: %w", err)
	}

	return &ChainCaller{
		taskMailbox: taskMailbox,
		chainId:     chainId,
		ethclient:   ethclient,
		logger:      logger,
		signer:      signer,
	}, nil
}

func (cc *ChainCaller) PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*types.Receipt, error) {
	opts, err := cc.signer.GetTransactOpts(ctx)
	if err != nil {
		return nil, err
	}

	address := cc.signer.GetFromAddress()

	tx, err := cc.taskMailbox.CreateTask(opts, ITaskMailbox.ITaskMailboxTypesTaskParams{
		RefundCollector: address,
		ExecutorOperatorSet: ITaskMailbox.OperatorSet{
			Avs: common.HexToAddress(avsAddress),
			Id:  operatorSetId,
		},
		Payload: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	receipt, err := cc.signAndSendTransaction(ctx, tx, "PublishMessageToInbox")
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully published message to inbox",
		zap.String("transactionHash", receipt.TxHash.Hex()),
	)
	return receipt, nil
}

//func (cc *ChainCaller) buildTransactionOpts(ctx context.Context) (*types.Transaction, error) {
//	// Get nonce
//	nonce, err := cc.ethclient.PendingNonceAt(ctx, cc.signer.GetFromAddress())
//	if err != nil {
//		return nil, fmt.Errorf("failed to get nonce: %w", err)
//	}
//
//	// Get gas price
//	gasPrice, err := cc.ethclient.SuggestGasPrice(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get gas price: %w", err)
//	}
//
//	// Create unsigned transaction
//	tx := types.NewTx(&types.LegacyTx{
//		Nonce:    nonce,
//		GasPrice: gasPrice,
//		Gas:      3000000, // Gas limit
//	})
//
//	return tx, nil
//}

func (cc *ChainCaller) signAndSendTransaction(ctx context.Context, tx *types.Transaction, operation string) (*types.Receipt, error) {
	// Sign the transaction
	signedTx, err := cc.signer.SignAndSendTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Wait for receipt
	receipt, err := cc.waitForReceipt(ctx, signedTx.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt, fmt.Errorf("transaction failed with status %d", receipt.Status)
	}

	return receipt, nil
}

func (cc *ChainCaller) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := cc.ethclient.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
			// Retry
		}
	}
}
