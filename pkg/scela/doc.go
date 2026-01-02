// Package scela provides an in-process message bus for Go applications.
//
// Scéla (Old Irish: "news, tidings, messages") enables pub/sub messaging patterns
// with support for synchronous and asynchronous delivery, pattern matching,
// middleware, retry logic, and dead letter queues.
//
// Basic usage:
//
//	bus := scela.New()
//	defer bus.Close()
//
//	// Subscribe to messages
//	bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
//	    fmt.Printf("User created: %v\n", msg.Payload())
//	    return nil
//	}))
//
//	// Publish message asynchronously
//	bus.Publish(context.Background(), "user.created", userData)
//
//	// Publish message synchronously
//	bus.PublishSync(context.Background(), "user.created", userData)
package scela
