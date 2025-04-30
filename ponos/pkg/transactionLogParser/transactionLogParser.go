package transactionLogParser

import (
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/convert"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// TransactionLogParser handles the parsing and decoding of Ethereum transaction logs.
// It uses contract ABIs to decode event data into structured format.
type TransactionLogParser struct {
	abi    *abi.ABI
	logger *zap.Logger
}

// NewTransactionLogParser creates a new TransactionLogParser with the provided dependencies.
//
// Parameters:
//   - logger: Logger for recording operations
//   - interestingLogQualifier: Qualifier to determine which logs to process
//
// Returns:
//   - *TransactionLogParser: A configured transaction log parser
func NewTransactionLogParser(abi *abi.ABI, logger *zap.Logger) *TransactionLogParser {
	return &TransactionLogParser{
		abi:    abi,
		logger: logger,
	}
}

func (tlp *TransactionLogParser) ProcessLog(
	log *ethereum.EthereumEventLog,
	block *ethereum.EthereumBlock,
	inboxAddress string,
	chainId config.ChainId,
) (*types.Task, error) {
	decoded, err := tlp.decodeLog(log)
	if err != nil {
		return nil, err
	}

	return convert.ConvertTask(decoded, block, inboxAddress, chainId)
}

// DecodeLog decodes a log using the provided ABI.
// It extracts the event name, arguments, and output data from the log.
// Returns the decoded log with structured event data and any error encountered during decoding.
// If no ABI is provided, returns an error.
//
// Parameters:
//   - a: The ABI to use for decoding
//   - lg: The log to decode
//
// Returns:
//   - *parser.DecodedLog: The decoded log with structured data
//   - error: Any error encountered during decoding
func (tlp *TransactionLogParser) decodeLog(lg *ethereum.EthereumEventLog) (*log.DecodedLog, error) {
	tlp.logger.Sugar().Infow(fmt.Sprintf("Decoding log with txHash: '%s' address: '%s'", lg.TransactionHash.Value(), lg.Address.Value()))
	logAddress := common.HexToAddress(lg.Address.Value())

	topicHash := common.Hash{}
	if len(lg.Topics) > 0 {
		// Handle case where the log has no topics
		// Original tx this failed on: https://holesky.etherscan.io/tx/0x044213f3e6c0bfa7721a1b6cc42a354096b54b20c52e4c7337fcfee09db80d90#eventlog
		topicHash = common.HexToHash(lg.Topics[0].Value())
	}

	decodedLog := &log.DecodedLog{
		Address:  logAddress.String(),
		LogIndex: lg.LogIndex.Value(),
	}

	if tlp.abi == nil {
		tlp.logger.Sugar().Errorw("No ABI provided for decoding log",
			zap.String("address", logAddress.String()),
		)
		return nil, errors.New("no ABI provided for decoding log")
	}

	event, err := tlp.abi.EventByID(topicHash)
	if err != nil {
		tlp.logger.Sugar().Debugw(fmt.Sprintf("Failed to find event by ID '%s'", topicHash))
		return decodedLog, err
	}

	decodedLog.EventName = event.RawName
	decodedLog.Arguments = make([]log.Argument, len(event.Inputs))

	for i, input := range event.Inputs {
		decodedLog.Arguments[i] = log.Argument{
			Name:    input.Name,
			Type:    input.Type.String(),
			Indexed: input.Indexed,
		}
	}

	if len(lg.Topics) > 1 {
		for i, param := range lg.Topics[1:] {
			d, err := parseLogValueForType(event.Inputs[i], param.Value())
			if err != nil {
				tlp.logger.Sugar().Errorw("Failed to parse log value for type", zap.Error(err))
			} else {
				decodedLog.Arguments[i].Value = d
			}
		}
	}

	if len(lg.Data) > 0 {
		// strip the leading 0x
		byteData, err := hex.DecodeString(lg.Data.Value()[2:])
		if err != nil {
			tlp.logger.Sugar().Errorw("Failed to decode data to bytes: ", err)
			return decodedLog, err
		}

		outputDataMap := make(map[string]interface{})
		err = tlp.abi.UnpackIntoMap(outputDataMap, event.Name, byteData)
		if err != nil {
			tlp.logger.Sugar().Errorw("Failed to unpack data",
				zap.Error(err),
				zap.String("hash", lg.TransactionHash.Value()),
				zap.String("address", lg.Address.Value()),
				zap.String("eventName", event.Name),
				zap.String("transactionHash", lg.TransactionHash.Value()),
			)
			return nil, errors.New("failed to unpack data")
		}

		decodedLog.OutputData = outputDataMap
	}
	return decodedLog, nil
}

// ParseLogValueForType converts an Ethereum log value to an appropriate Go type
// based on the ABI argument type.
// It handles integer, boolean, address, string, and byte types.
//
// Parameters:
//   - argument: The ABI argument definition containing type information
//   - value: The hex-encoded value to parse
//
// Returns:
//   - interface{}: The converted value
//   - error: Any error encountered during conversion
func parseLogValueForType(argument abi.Argument, value string) (interface{}, error) {
	valueBytes, _ := hexutil.Decode(value)
	switch argument.Type.T {
	case abi.IntTy, abi.UintTy:
		return abi.ReadInteger(argument.Type, valueBytes)
	case abi.BoolTy:
		return readBool(valueBytes)
	case abi.AddressTy:
		return common.HexToAddress(value), nil
	case abi.StringTy:
		return value, nil
	case abi.BytesTy, abi.FixedBytesTy:
		// return value as-is; hex encoded string
		return value, nil
	default:
		return value, nil
	}
}

// errBadBool is returned when a boolean value in an Ethereum log is improperly encoded.
var (
	errBadBool = fmt.Errorf("abi: improperly encoded boolean value")
)

// readBool converts a 32-byte word to a boolean value.
// Valid encodings have all bytes except the last one set to zero,
// and the last byte set to either 0 (false) or 1 (true).
//
// Parameters:
//   - word: The 32-byte array to convert
//
// Returns:
//   - bool: The decoded boolean value
//   - error: Error if the encoding is invalid
func readBool(word []byte) (bool, error) {
	for _, b := range word[:31] {
		if b != 0 {
			return false, errBadBool
		}
	}
	switch word[31] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errBadBool
	}
}
