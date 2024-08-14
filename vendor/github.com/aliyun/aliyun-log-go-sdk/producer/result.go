package producer

type Attempt struct {
	Success      bool
	RequestId    string
	ErrorCode    string
	ErrorMessage string
	TimeStampMs  int64
}

func createAttempt(success bool, requestId, errorCode, errorMessage string, timeStampMs int64) *Attempt {
	return &Attempt{
		Success:      success,
		RequestId:    requestId,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		TimeStampMs:  timeStampMs,
	}
}

type Result struct {
	attemptList []*Attempt
	successful  bool
}

func (result *Result) IsSuccessful() bool {
	return result.successful
}

func (result *Result) GetReservedAttempts() []*Attempt {
	return result.attemptList
}

func (result *Result) GetErrorCode() string {
	if len(result.attemptList) == 0 {
		return ""
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].ErrorCode
}

func (result *Result) GetErrorMessage() string {
	if len(result.attemptList) == 0 {
		return ""
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].ErrorMessage
}

func (result *Result) GetRequestId() string {
	if len(result.attemptList) == 0 {
		return ""
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].RequestId
}

func (result *Result) GetTimeStampMs() int64 {
	if len(result.attemptList) == 0 {
		return 0
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].TimeStampMs
}

func initResult() *Result {
	return &Result{
		attemptList: []*Attempt{},
		successful:  false,
	}
}
