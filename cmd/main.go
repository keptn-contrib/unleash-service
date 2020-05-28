package main

import (
	"context"
	"log"
	"os"

	"github.com/keptn-contrib/unleash-service/pkg/event_handler"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/kelseyhightower/envconfig"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/keptn/go-utils/pkg/lib"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	os.Exit(_main(os.Args[1:], env))
}

func _main(args []string, env envConfig) int {

	ctx := context.Background()

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithPort(env.Port),
		cloudeventshttp.WithPath(env.Path),
	)

	log.Printf("Port = %d; Path=%s", env.Port, env.Path)

	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))

	return 0
}

func gotEvent(ctx context.Context, event cloudevents.Event) error {

	var shkeptncontext string
	_ = event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	logger := keptn.NewLogger(shkeptncontext, event.Context.GetID(), "unleash-service")

	eventHandler, err := event_handler.NewEventHandler(event, logger)

	if err != nil {
		return err
	}

	if eventHandler == nil {
		logger.Debug("No event handler registered for: " + event.Type())
		return nil
	}

	err = eventHandler.HandleEvent()

	if err != nil {
		return err
	}

	return nil
}
