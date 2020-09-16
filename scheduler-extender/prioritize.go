package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

type Prioritize struct {
	Name string
	Func func(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error)
}

func (p Prioritize) Handler(args schedulerapi.ExtenderArgs) (*schedulerapi.HostPriorityList, error) {
	return p.Func(*args.Pod, args.Nodes.Items)
}

func PrioritizeRoute(prioritize Prioritize) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)

		var extenderArgs schedulerapi.ExtenderArgs
		var hostPriorityList *schedulerapi.HostPriorityList

		if err := json.NewDecoder(body).Decode(&extenderArgs); err != nil {
			panic(err)
		}
		log.Print(prioritize.Name, " ExtenderArgs = ", buf.String())

		if list, err := prioritize.Handler(extenderArgs); err != nil {
			panic(err)
		} else {
			hostPriorityList = list
		}

		for _, hostPriority := range *hostPriorityList {
			if hostPriority.Score < schedulerapi.MinExtenderPriority || hostPriority.Score > schedulerapi.MaxExtenderPriority {
				panic(fmt.Errorf("invalid score: %v(%v)", hostPriority.Score, hostPriority.Host))
			}
		}

		if resultBody, err := json.Marshal(hostPriorityList); err != nil {
			panic(err)
		} else {
			log.Print(prioritize.Name, " hostPriorityList = ", string(resultBody))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}
