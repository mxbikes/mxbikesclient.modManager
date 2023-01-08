package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/mxbikes/mxbikesclient.modManager/client"
	"github.com/mxbikes/mxbikesclient.modManager/models"
	"github.com/mxbikes/mxbikesclient.modManager/repository"
	"github.com/sirupsen/logrus"
	"gopkg.in/toast.v1"
)

type Events struct {
	repository repository.SubscriptionRepository
	logger     logrus.Logger
	modService client.ModServiceClient
	userID     string
}

const log_withEventAggregateID = "subscription with id: {%s} "

// Return a new handler
func NewEventHandler(filStoreDB repository.SubscriptionRepository, logger logrus.Logger, modService client.ModServiceClient, userID string) *Events {
	return &Events{repository: filStoreDB, logger: logger, modService: modService, userID: userID}
}

func (o *Events) OnSubscriptionAdded(ctx context.Context, evt *esdb.RecordedEvent) error {
	var aggregateId = strings.ReplaceAll(evt.StreamID, "subscription-", "")

	if o.userID != aggregateId {
		return errors.New("Not this account")
	}

	// Get data from event
	var subscription models.Subscription
	if err := json.Unmarshal(evt.Data, &subscription); err != nil {
		return err
	}

	o.logger.WithFields(logrus.Fields{"prefix": "SERVICE.CommentEvent_OnSubscriptionAdded"}).Infof(log_withEventAggregateID, aggregateId)

	// Get mod information
	mod, err := o.modService.GetModByID(subscription.ModID)
	if err != nil {
		return err
	}

	// Does file exist
	fileExist := o.repository.DoesFileExist(mod.Mod.Name)

	// Download mod
	err = o.repository.DownloadModFile(subscription.ModID, mod.Mod.Name)
	if err != nil {
		fmt.Println(err)
	}

	if fileExist {
		notification := toast.Notification{
			Icon:    `C:\Users\Developer\Source\repos\mxbikesclient.modManager\utils\logo.png`,
			AppID:   "MxBikesClient",
			Title:   fmt.Sprintf("Mod: %s", mod.Mod.Name),
			Message: "The mod has been updated successfully!",
		}
		err = notification.Push()
		if err != nil {
			log.Fatalln(err)
		}

		return nil
	}

	// Toast notification
	notification := toast.Notification{
		Icon:    `C:\Users\Developer\Source\repos\mxbikesclient.modManager\utils\logo.png`,
		AppID:   "MxBikesClient",
		Title:   fmt.Sprintf("Mod: %s", mod.Mod.Name),
		Message: "The mod has been installed successfully!",
	}
	err = notification.Push()
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (o *Events) OnSubscriptionRemoved(ctx context.Context, evt *esdb.RecordedEvent) error {
	var aggregateId = strings.ReplaceAll(evt.StreamID, "subscription-", "")

	if o.userID != aggregateId {
		return errors.New("Not this account")
	}

	// Get data from event
	var subscription models.Subscription
	if err := json.Unmarshal(evt.Data, &subscription); err != nil {
		return err
	}

	o.logger.WithFields(logrus.Fields{"prefix": "SERVICE.CommentEvent_OnSubscriptionAdded"}).Infof(log_withEventAggregateID, aggregateId)

	// Get mod information
	mod, err := o.modService.GetModByID(subscription.ModID)
	if err != nil {
		return err
	}

	// Does file exist
	fileExist := o.repository.DoesFileExist(mod.Mod.Name)
	if !fileExist {
		return errors.New("file doesnt exist")
	}

	// Delete mod
	err = o.repository.DeleteModFile(mod.Mod.Name)
	if err != nil {
		fmt.Println(err)
	}

	// Toast notification
	notification := toast.Notification{
		Icon:    `C:\Users\Developer\Source\repos\mxbikesclient.modManager\utils\logo.png`,
		AppID:   "MxBikesClient",
		Title:   fmt.Sprintf("Mod: %s", mod.Mod.Name),
		Message: "The mod has been removed successfully!",
	}
	err = notification.Push()
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}
