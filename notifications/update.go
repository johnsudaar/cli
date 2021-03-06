package notifications

import (
	"gopkg.in/errgo.v1"
	"github.com/Scalingo/cli/config"
	"github.com/Scalingo/cli/io"
)

func Update(app, ID, webHookURL string) error {
	if app == "" {
		return errgo.New("no app defined")
	} else if webHookURL == "" {
		return errgo.New("no url defined")
	}

	_, err := checkNotificationExist(app, ID)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	c := config.ScalingoClient()
	params, err := c.NotificationUpdate(app, ID, webHookURL)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	io.Status("Notifications are now sent to", webHookURL)
	if len(params.Variables) > 0 {
		io.Info("Modified variables:", params.Variables)
	}
	if len(params.Message) > 0 {
		io.Info("Message from notification updater:", params.Message)
	}
	return nil
}
