package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kube-scheduler/extender/v1"
)

var (
	apiPrefix      = "/api/scheduler"
	predicatesPath = apiPrefix + "/predicates"
	prioritiesPath = apiPrefix + "/priorities"
)

func main() {
	router := httprouter.New()

	predicate := Predicate{
		Name: "always-true",
		Func: alwaysTrue,
	}
	router.POST(predicatesPath, PredicateRoute(predicate))

	prioritize := Prioritize{
		Name: "zero-score",
		Func: zeroScore,
	}
	router.POST(prioritiesPath, PrioritizeRoute(prioritize))

	log.Print("server starting on the port :12345")
	if err := http.ListenAndServe(":12345", router); err != nil {
		log.Fatal(err)
	}
}

func alwaysTrue(pod v1.Pod, node v1.Node) (bool, error) {
	return true, nil
}

func zeroScore(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
	var priorityList schedulerapi.HostPriorityList
	priorityList = make([]schedulerapi.HostPriority, len(nodes))
	for i, node := range nodes {
		priorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: 0,
		}
	}
	return &priorityList, nil
}
