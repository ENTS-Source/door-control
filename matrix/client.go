package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

var client *mautrix.Client

func Connect(homeserverUrl string, userId string, accessToken string) error {
	Stop()

	var err error
	client, err = mautrix.NewClient(homeserverUrl, id.UserID(userId), accessToken)
	if err != nil {
		return err
	}

	return nil
}

func Stop() {
	// Nothing to do
}
