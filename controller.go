package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	cron "gopkg.in/robfig/cron.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller struct defines how a controller should encapsulate
// logging, client connectivity, informing (list and watching)
// queueing, and handling of resource changes
type Controller struct {
	logger    *log.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	handler   Handler
	croner    *cron.Cron
	jober     map[string]cron.EntryID
}

// Run is the main path of execution for the controller loop
func (c *Controller) Run(stopCh <-chan struct{}) {
	// handle a panic with logging and exiting
	defer utilruntime.HandleCrash()
	// ignore new items in the queue but when all goroutines
	// have completed existing items then shutdown
	defer c.queue.ShutDown()

	c.logger.Info("Controller.Run: initiating")

	// run the informer to start listing and watching resources
	go c.informer.Run(stopCh)

	// do the initial synchronization (one time) to populate resources
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	c.logger.Info("Controller.Run: cache sync complete")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced allows us to satisfy the Controller interface
// by wiring up the informer's HasSynced method to it
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// runWorker executes the loop to process new items added to the queue
func (c *Controller) runWorker() {

	log.Info("Controller.runWorker: starting")

	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource

	for c.processNextItem() {
		log.Info("Controller.runWorker: processing next item")
	}
	log.Info("Controller.runWorker: completed")
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *Controller) processNextItem() bool {
	log.Info("Controller.processNextItem: start")

	// fetch the next item (blocking) from the queue to process or
	// if a shutdown is requested then return out of this to stop
	// processing
	key, quit := c.queue.Get()

	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the queue has indicated
	// from the Get method
	if quit {
		return false
	}

	defer c.queue.Done(key)

	// assert the string out of the key (format `namespace/name`)
	keyRaw := key.(string)

	// take the string key and get the object out of the indexer
	//
	// item will contain the complex object for the resource and
	// exists is a bool that'll indicate whether or not the
	// resource was created (true) or deleted (false)
	//
	// if there is an error in getting the key from the index
	// then we want to retry this particular queue key a certain
	// number of times (5 here) before we forget the queue key
	// and throw an error
	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(key) < 5 {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
			c.queue.AddRateLimited(key)
		} else {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", key, err)
			c.queue.Forget(key)
			utilruntime.HandleError(err)
		}
	}

	// if the item doesn't exist then it was deleted and we need to fire off the handler's
	// ObjectDeleted method. but if the object does exist that indicates that the object
	// was created (or updated) so run the ObjectCreated method
	//
	// after both instances, we want to forget the key from the queue, as this indicates
	// a code path of successful queue key processing
	if !exists {
		c.logger.Infof("Controller.processNextItem: object deleted detected: %s", keyRaw)
		c.handler.ObjectDeleted(item)
		c.queue.Forget(key)

		id := c.jober[keyRaw]
		c.croner.Remove(id)

	} else {
		c.logger.Infof("Controller.processNextItem: object created detected: %s", keyRaw)
		namespace, schedule := c.handler.ObjectCreated(item)

		// TO DO: Check if namespace exists

		// TO DO: default schedule: "0 * * * * *"

		switch namespace {
		//Restart node
		case "nodes":
			id, _ := c.croner.AddFunc(schedule, func() {
				//cmd := exec.Command("/bin/systemctl", "reboot")
				cmd := exec.Command("/sbin/shutdown", "-r", "now")
				c.logger.Infof("Running restart node command ...")
				err := cmd.Run()
				c.logger.Infof("Command finished with error: %v", err)
			})

			c.jober[keyRaw] = id
			c.queue.Forget(key)
		// Delete pod in the given namespace
		default:
			id, _ := c.croner.AddFunc(schedule, func() {

				pods, _ := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
				p := rand.Int() % len(pods.Items)
				err = c.clientset.CoreV1().Pods(namespace).Delete(pods.Items[p].Name, nil)

				if err != nil {
					log.WithFields(log.Fields{
						"namespace": namespace,
						"pod":       pods.Items[p].Name,
					}).Fatal("Error deleting pod")
				} else {
					log.WithFields(log.Fields{
						"namespace": namespace,
						"pod":       pods.Items[p].Name,
					}).Info("Pod deleted")
				}
			})

			c.jober[keyRaw] = id
			c.queue.Forget(key)
		}
	}

	return true
}
