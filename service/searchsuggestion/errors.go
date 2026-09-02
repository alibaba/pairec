package searchsuggestion

import "fmt"

const (
	CodeConfigUnavailable         = "suggestion_config_unavailable"
	CodeKnowledgeUnavailable      = "suggestion_knowledge_unavailable"
	CodeKnowledgeFailed           = "suggestion_knowledge_failed"
	CodeKnowledgeEmpty            = "suggestion_knowledge_empty"
	CodePromptMissing             = "suggestion_prompt_missing"
	CodeModelUnavailable          = "suggestion_model_unavailable"
	CodeProductContextUnavailable = "suggestion_product_context_unavailable"
	CodeProductSearchFailed       = "suggestion_product_search_failed"
	CodeProductContextInvalid     = "suggestion_product_context_invalid"
	CodeMainReplyIneligible       = "suggestion_main_reply_ineligible"
	CodeNotReady                  = "suggestion_not_ready"
	CodeTimeout                   = "suggestion_timeout"
	CodeModelError                = "suggestion_model_error"
	CodeInvalidStructure          = "suggestion_invalid_structure"
	CodeValidationFailed          = "suggestion_validation_failed"
)

type Error struct {
	Code      string
	Retryable bool
	Cause     error
}

func NewError(code string, retryable bool, cause error) *Error {
	return &Error{Code: code, Retryable: retryable, Cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Cause)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) PublicMessage(language string) string {
	switch language {
	case "zh":
		return "追问建议暂时不可用"
	case "ar":
		return "اقتراحات البحث غير متاحة مؤقتًا"
	default:
		return "Search suggestions are temporarily unavailable"
	}
}

type Outcome struct {
	Suggestions []string
	Err         *Error
}
