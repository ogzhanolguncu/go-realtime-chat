package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/ogzhanolguncu/go-chat/client/internal"
	ui_manager "github.com/ogzhanolguncu/go-chat/client/ui_manager/handlers"
)

func main() {
	retry.Do(
		func() error {
			return runClient()
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			if err.Error() == io.EOF.Error() {
				err = fmt.Errorf("server is not responding")
			}
			fmt.Printf("Trying to reconnect, but %v\n", err)
		}),
	)
}

func runClient() error {
	client, err := internal.NewClient(internal.NewConfig())
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect server: %v", err)
	}

	if err := manageUIs(client); err != nil {
		log.Fatalf("Error managing UIs: %v", err)
	}
	return nil
}

func manageUIs(client *internal.Client) error {
	terminate, err := ui_manager.HandleLoginUI(client)
	if terminate {
		return nil
	}
	if err != nil {
		return err
	}
	for {
		switchToChannel, err := ui_manager.HandleChatUI(client)
		if err != nil {
			return fmt.Errorf("error in chat UI: %v", err)
		}

		if switchToChannel {
			if err := ui_manager.HandleChannelUI(client); err != nil {
				return fmt.Errorf("error in channel UI: %v", err)
			}
		} else {
			// User wants to quit the application
			return nil
		}
	}
}
