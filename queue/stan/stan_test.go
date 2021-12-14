// +build integration

package stan

import (
	"github.com/nats-io/stan.go"
	"log"
	"sync"
	"testing"
	"time"
)

func Test_SyncConnect(t *testing.T) {
	sc, err := stan.Connect("test-cluster", "test_client")
	if err != nil {
		t.Fatal(err)
	}

	err = sc.Publish("test-subj", []byte("Test"))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := sc.Subscribe("test-subj", func(m *stan.Msg) {
		log.Printf("received: %v", m)
	}, stan.StartWithLastReceived())
	if err != nil {
		t.Fatal(err)
	}

	err = sub.Unsubscribe()
	if err != nil {
		t.Fatal(err)
	}

	err = sc.Close()
	if err != nil {
		t.Fatal(err)
	}

}

func initDurable(t *testing.T, topic string) {

	sb, err := stan.Connect("test-cluster", "init")
	if err != nil {
		t.Fatal(err)
	}
	defer sb.Close()
	s, err := sb.Subscribe(topic, func(m *stan.Msg) {
		log.Printf("received: %v\n", m)
	},  stan.DurableName("durable"))
	if err != nil {
		t.Fatal(err)
	}
	err = s.Unsubscribe()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Durable(t *testing.T) {

	subj := "subj"
	//initDurable(t, subj)

	sendChan := make(chan string)
	errChan := make(chan error)
	quitPublisher := make(chan interface{})
	quitSubscriber := make(chan interface{})

	var wg sync.WaitGroup

	go func(){
		t.Fatal(<-errChan)
	}()

	// publisher
	go func() {
		pb, err := stan.Connect("test-cluster", "publisher")
		if err != nil {
			errChan <- err
			return
		}
		defer pb.Close()
		for {
			select {
				case msg := <- sendChan:
					log.Printf("send: %s", msg)
					err := pb.Publish(subj, []byte(msg))
					if err != nil {
						errChan <- err
					}
				case <- quitPublisher: return
			}
		}

	}()

	// subscriber
	subscriber := func() {
		sb, err := stan.Connect("test-cluster", "subscriber")
		if err != nil {
			errChan <- err
			return
		}
		defer sb.Close()
		_, err = sb.Subscribe(subj, func(m *stan.Msg) {
			log.Printf("received: %v\n", m)
			wg.Done()
		},  stan.DurableName("durable"))
		if err != nil {
			errChan <- err
			return
		}
		//defer subscription.Unsubscribe()
		<- quitSubscriber
		log.Println("subscriber closed")
	}

	wg.Add(2)
	sendChan <- "msg-1"
	sendChan <- "msg-2"
	log.Println("run subscriber")
	go subscriber()
	wg.Wait()
	log.Println("sending quit subscriber")
	quitSubscriber <- true
	wg.Add(2)
	sendChan <- "msg-3"
	sendChan <- "msg-4"
	log.Println("run subscriber")
	go subscriber()

	log.Println("waiting for wg")
	wg.Wait()

	log.Println("quit all")
	quitSubscriber <- true
	quitPublisher <- true

}

func Test_Queue(t *testing.T) {

	initDurable(t, "test-queue")

	sendChan := make(chan string)
	errChan := make(chan error)
	quitPublisher := make(chan interface{})
	quitSubscriber := make(chan interface{})

	var wg sync.WaitGroup

	go func(){
		t.Fatal(<-errChan)
	}()

	// publisher
	go func() {
		pb, err := stan.Connect("test-cluster", "publisher")
		if err != nil {
			errChan <- err
			return
		}
		defer pb.Close()
		for {
			select {
			case msg := <- sendChan:
				log.Printf("send: %s", msg)
				err := pb.Publish("test-queue", []byte(msg))
				if err != nil {
					errChan <- err
				}
			case <- quitPublisher: return
			}
		}

	}()

	// subscriber
	subscriber := func(id string, quitChan <- chan interface{}) {
		sb, err := stan.Connect("test-cluster", "subscriber" + id)
		if err != nil {
			errChan <- err
			return
		}
		defer sb.Close()
		_, err = sb.QueueSubscribe("test-queue", "queue", func(m *stan.Msg) {
			log.Printf("received (%s): %v\n", id, m)
			wg.Done()
		},  stan.DurableName("durable"))
		if err != nil {
			errChan <- err
			return
		}
		//defer subscription.Unsubscribe()
		<- quitChan
		log.Printf("subscriber %s closed", id)
	}

	log.Println("run subscribers")
	go subscriber("1", quitSubscriber)
	go subscriber("2", quitSubscriber)

	time.Sleep(time.Second)

	wg.Add(2)
	sendChan <- "msg-1"
	time.Sleep(time.Millisecond * 10)
	sendChan <- "msg-2"

	wg.Wait()
	log.Println("sending quit subscriber")
	quitSubscriber <- true
	quitSubscriber <- true
	wg.Add(2)
	sendChan <- "msg-3"
	time.Sleep(time.Millisecond * 10)
	sendChan <- "msg-4"
	log.Println("run subscriber")
	go subscriber("3",quitSubscriber)
	go subscriber("4", quitSubscriber)

	log.Println("waiting for wg")
	wg.Wait()

	log.Println("quit all")
	quitSubscriber <- true
	quitSubscriber <- true
	quitPublisher <- true

}