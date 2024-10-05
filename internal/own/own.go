package own

import (
	"github.com/pkg/errors"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/owu-one/gotosocial-sdk/models"
)

// Account returns the currently authenticated account.
func Account(authClient *auth.Client) (*models.Account, error) {
	err := authClient.Wait()
	if err != nil {
		return nil, err
	}

	resp, err := authClient.Client.Accounts.AccountVerify(nil, authClient.Auth)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.GetPayload(), nil
}

// Instance returns the instance of the currently authenticated account.
func Instance(authClient *auth.Client) (*models.InstanceV2, error) {
	err := authClient.Wait()
	if err != nil {
		return nil, err
	}

	resp, err := authClient.Client.Instance.InstanceGetV2(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.GetPayload(), nil
}

func Domain(authClient *auth.Client) (string, error) {
	ownInstance, err := Instance(authClient)
	if err != nil {
		return "", err
	}

	ownDomain := ownInstance.AccountDomain
	if ownDomain == "" {
		ownDomain = ownInstance.Domain
	}
	if ownDomain == "" {
		return "", errors.WithStack(errors.New("couldn't find domain for accounts on this instance"))
	}

	return ownDomain, nil
}
