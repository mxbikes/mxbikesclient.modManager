package projection

import (
	"context"
	"errors"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/mxbikes/mxbikesclient.modManager/client"
	"github.com/mxbikes/mxbikesclient.modManager/handler"
	"github.com/mxbikes/mxbikesclient.modManager/repository"
	"github.com/sirupsen/logrus"
)

type projection struct {
	db      *esdb.Client
	logger  logrus.Logger
	handler *handler.Events
	userID  string
}

// Return a new handler
func New(logger logrus.Logger, db *esdb.Client, fileRepository repository.SubscriptionRepository, modService client.ModServiceClient, userID string) *projection {
	return &projection{handler: handler.NewEventHandler(fileRepository, logger, modService, userID), logger: logger, db: db, userID: userID}
}

func (o *projection) Subscribe(ctx context.Context) error {
	o.logger.WithFields(logrus.Fields{"prefix": "POSTGRES_SUBSCRIPTION"}).Infof("prefixes: {%+v}", []string{"subscription-" + o.userID})
	err := o.db.CreatePersistentSubscriptionAll(context.Background(), "subscriptionClient", esdb.PersistentAllSubscriptionOptions{
		Filter: &esdb.SubscriptionFilter{Type: esdb.StreamFilterType, Prefixes: []string{"subscription-" + o.userID}},
	})
	if err != nil {
		if subscriptionError, ok := err.(*esdb.PersistentSubscriptionError); !ok || ok && (subscriptionError.Code != 6) {
			o.logger.WithFields(logrus.Fields{"prefix": "POSTGRES_SUBSCRIPTION"}).Errorf("failed createPersistentSubscriptionAll: err: {%v}", subscriptionError.Error())
		}
	}

	stream, err := o.db.ConnectToPersistentSubscription(
		ctx,
		"$all",
		"subscriptionClient",
		esdb.ConnectToPersistentSubscriptionOptions{},
	)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	o.ProcessEvents(ctx, stream)
	return nil
}

func (o *projection) ProcessEvents(ctx context.Context, stream *esdb.PersistentSubscription) error {
	for {
		event := stream.Recv()

		if event.EventAppeared != nil {

			err := o.When(ctx, event.EventAppeared.Event)
			if err != nil {
				if err := stream.Nack(err.Error(), esdb.Nack_Retry, event.EventAppeared); err != nil {
					o.logger.WithFields(logrus.Fields{"prefix": "STREAM.NACk"}).Errorf("err: {%v}", err)
					return err
				}
			}

			err = stream.Ack(event.EventAppeared)
			if err != nil {
				o.logger.WithFields(logrus.Fields{"prefix": "STREAM.NACk"}).Errorf("err: {%v}", err)
				return err
			}
		}

		if event.SubscriptionDropped != nil {
			break
		}
	}

	return nil
}

func (o *projection) When(ctx context.Context, evt *esdb.RecordedEvent) error {
	switch evt.EventType {

	case "SUBSCRIPTION_ADDED":
		return o.handler.OnSubscriptionAdded(ctx, evt)
	case "SUBSCRIPTION_REMOVED":
		return o.handler.OnSubscriptionRemoved(ctx, evt)

	default:
		o.logger.WithFields(logrus.Fields{"prefix": "Projection"}).Warnf("unknown eventType: {%s}", evt.EventType)
		return errors.New("Unkown type")
	}
}

/*!SECTION

err = eventStoreDB.CreatePersistentSubscriptionAll(context.Background(), "subscriptionClient", esdb.PersistentAllSubscriptionOptions{
		Filter: &esdb.SubscriptionFilter{Type: esdb.StreamFilterType, Prefixes: []string{"subscription-"}},
	})
	if err != nil {
		if subscriptionError, ok := err.(*esdb.PersistentSubscriptionError); !ok || ok && (subscriptionError.Code != 6) {
			fmt.Errorf("(CreatePersistentSubscriptionAll) err: {%v}", subscriptionError.Error())
		}
	}

	stream, err := eventStoreDB.ConnectToPersistentSubscription(context.Background(), "$all", "subscriptionClient", esdb.ConnectToPersistentSubscriptionOptions{})
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	go func() {
		for {
			event := stream.Recv()

			if event.EventAppeared != nil {
				err = stream.Ack(event.EventAppeared)
				if err != nil {
					fmt.Errorf("(stream.Ack) err: {%v}", err)
				}
				fmt.Printf("(ACK) event commit: {%v}", *event.EventAppeared.Commit)
			}

			if event.SubscriptionDropped != nil {
				break
			}
		}
	}()


*/
