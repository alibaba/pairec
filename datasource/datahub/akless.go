package datahub

import (
	"fmt"

	alidatahub "github.com/aliyun/aliyun-datahub-sdk-go/datahub"
	"github.com/aliyun/credentials-go/credentials"
)

var _ alidatahub.Account = (*AklessAccount)(nil)

type AklessAccount struct {
	credential credentials.Credential
}

func NewAklessAccount() (*AklessAccount, error) {
	credential, err := credentials.NewCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create akless account: %v", err)
	}
	return &AklessAccount{credential: credential}, nil
}

// GetAccountId implements datahub.Account.
func (a *AklessAccount) GetAccountId() string {
	credntialModel, _ := a.credential.GetCredential()
	return *credntialModel.AccessKeyId
}

// GetAccountKey implements datahub.Account.
func (a *AklessAccount) GetAccountKey() string {
	credntialModel, _ := a.credential.GetCredential()
	return *credntialModel.AccessKeySecret
}

// GetSecurityToken implements datahub.Account.
func (a *AklessAccount) GetSecurityToken() string {
	credntialModel, _ := a.credential.GetCredential()
	return *credntialModel.SecurityToken
}

// String implements datahub.Account.
func (a *AklessAccount) String() string {
	credntialModel, _ := a.credential.GetCredential()
	return fmt.Sprintf("accessId: %s, accessKey: %s, stsToken:%s", *credntialModel.AccessKeyId, *credntialModel.AccessKeySecret, *credntialModel.SecurityToken)
}
