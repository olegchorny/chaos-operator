package main

import (
	verfv1 "github.com/olegchorny/chaos-operator/pkg/apis/chaos/v1"

	log "github.com/Sirupsen/logrus"
)

// Handler interface contains the methods that are required
type Handler interface {
	Init() error
	ObjectCreated(obj interface{}) (namespace string, schedule string)
	ObjectDeleted(obj interface{})
	ObjectUpdated(objOld, objNew interface{})
}

// ChaosHandler is a sample implementation of Handler
type ChaosHandler struct{}

// Init handles any handler initialization
func (t *ChaosHandler) Init() error {
	log.Info("ChaosHandler.Init")
	return nil
}

// ObjectCreated is called when an object is created
func (t *ChaosHandler) ObjectCreated(obj interface{}) (namespace string, schedule string) {
	log.Info("ChaosHandler.ObjectCreated")

	mr := obj.(*verfv1.Chaos)

	log.WithFields(log.Fields{
		"namespace": mr.Spec.Namespace,
		"schedule":  mr.Spec.Schedule,
	}).Info("new chaos is scheduled")
	return mr.Spec.Namespace, mr.Spec.Schedule
}

// ObjectDeleted is called when an object is deleted
func (t *ChaosHandler) ObjectDeleted(obj interface{}) {
	log.Info("ChaosHandler.ObjectDeleted")
}

// ObjectUpdated is called when an object is updated
func (t *ChaosHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Info("ChaosHandler.ObjectUpdated")
}
