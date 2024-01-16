package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	seeds := []string{"127.0.0.1:19092"}
	// One client can both produce and consume!
	// Consuming can either be direct (no consumer group), or through a group. Below, we use a group.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("my-group-identifier"),
		kgo.ConsumeTopics("foo"),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	ctx := context.Background()

	// 1. Producing a message
	// All record production goes through Produce, and the callback can be used
	// to allow for synchronous or asynchronous production.
	var wg sync.WaitGroup

	record := &kgo.Record{
		Topic: "foo",
		Value: []byte("bar"),
	}

	wg.Add(1)
	cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()

		if err != nil {
			log.Printf("record had a produce error: %v\n", err)
		}
	})
	log.Printf("waiting for event to send")
	wg.Wait()

	// Alternatively, ProduceSync exists to synchronously produce a batch of records.
	if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Printf("record had a produce error while synchronously producing: %v\n", err)
	}

	log.Printf("two new records are produced using synchronous and asynchronous manners\n")

	// 2. Consuming messages from a topic
	for {
		log.Printf("waiting for new records to come...")

		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			// All errors are retried internally when fetching, but non-retriable errors are
			// returned from polls so that users can notice and take action.
			panic(fmt.Sprint(errs))
		}

		// We can iterate through a record iterator...
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			log.Println(string(record.Value), "from an iterator!")
		}

		// or a callback function.
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				log.Println(string(record.Value), "from range inside a callback!")
			}

			// We can even use a second callback!
			p.EachRecord(func(record *kgo.Record) {
				log.Println(string(record.Value), "from a second callback!")
			})
		})
	}
}
