package rank

import "math"

// DefaultDebugLevel is the debug_level the rank and general rank paths have
// always sent to the easyrec processor for a debug request.
const DefaultDebugLevel = 100

// ResolveDebugLevel returns the debug_level of a debug request. A positive
// request level overrides the default one, any other value (unset, non
// positive or out of the int32 range of PBRequest.debug_level) keeps the
// default so the behaviour stays the same as before.
func ResolveDebugLevel(requestLevel int) int {
	if requestLevel > 0 && requestLevel <= math.MaxInt32 {
		return requestLevel
	}

	return DefaultDebugLevel
}
