package shoppingknowledge

import (
	"fmt"

	"github.com/alibaba/pairec/v2/log"
)

// LogModelView logs the same knowledge JSON sent to the model.
func (e *Evidence) LogModelView(requestID string) {
	if e == nil {
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ShoppingKnowledge\tevent=model_view\tpayload=%s",
		requestID, e.PromptJSON()))
}
