package main

import "github.com/TheMeier/k8sinfo/model"

// K8sInfoChannels provides get/set channels for syncronization
type k8sInfoChannelsHolder struct {
	set chan model.K8sInfoData
	get chan model.K8sInfoData
}

func (h k8sInfoChannelsHolder) mux() {
	var data model.K8sInfoData
	for {
		select {
		case data = <-h.set: // set the current value.
		case h.get <- data: // send the current value.
		}
	}
}

// K8sInfoHolder holds K8sInfoData in a concurrency-safe manner
type K8sInfoHolder interface {
	Get() model.K8sInfoData
	Set(model.K8sInfoData)
}

func (h k8sInfoChannelsHolder) Get() model.K8sInfoData {
	return <-h.get
}

func (h k8sInfoChannelsHolder) Set(s model.K8sInfoData) {
	h.set <- s
}

// NewK8sInfoHolder returns a helper object holding the channels
func NewK8sInfoHolder() K8sInfoHolder {
	h := k8sInfoChannelsHolder{
		set: make(chan model.K8sInfoData),
		get: make(chan model.K8sInfoData),
	}
	go h.mux()
	return h
}
