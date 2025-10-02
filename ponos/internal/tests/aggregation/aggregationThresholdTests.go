package aggregation

const (
	l1RpcUrl                = "http://127.0.0.1:8545"
	numExecutorOperators    = 4
	aggregatorOperatorSetId = 0
	executorOperatorSetId   = 1
	maxStalenessPeriod      = 604800
	transportBlsKey         = "0x5f8e6420b9cb0c940e3d3f8b99177980785906d16fb3571f70d7a05ecf5f2172"
)

type thresholdTestCase struct {
	name                       string
	aggregationThreshold       uint16
	verificationThreshold      uint16
	respondingOperatorIdxs     []int
	shouldVerifySucceed        bool
	shouldMeetSigningThreshold bool
	operatorResponses          map[int][]byte
}
