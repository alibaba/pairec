package rank

// DefaultDebugLevel is the debug_level the rank and general rank paths have
// always sent to the easyrec processor for a debug request, it saves the
// prediction request on the processor side.
const DefaultDebugLevel = 100

// ResolveDebugLevel returns the debug_level of a debug request. Only the levels
// documented by the easyrec processor are passed through: 1 and 3 return the
// features without predicting, 2 returns the features and tensors and predicts,
// 100 saves the prediction request, 102 is the vector recall debug level, 903
// prints the prediction time of every stage and 904 logs the missing feature
// fields. Any other value, including the unset 0, keeps DefaultDebugLevel.
func ResolveDebugLevel(requestLevel int) int {
	switch requestLevel {
	case 1, 2, 3, 100, 102, 903, 904:
		return requestLevel
	default:
		return DefaultDebugLevel
	}
}
