package main

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func getNewLock(lockname, podname, namespace string) *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockname,
			Namespace: namespace,
		},
		Client: config.KubernetesClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podname,
		},
	}
}

func setLeaderLabel(isLeader bool) {
	pod, err := config.KubernetesClient.CoreV1().Pods(namespace).Get(context.TODO(), hostname, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	existingLabels := pod.ObjectMeta.Labels
	if isLeader {
		existingLabels["leader"] = "true"
	} else {
		existingLabels["leader"] = "false"
	}

	pod.ObjectMeta.Labels = existingLabels

	updatedPod, err := config.KubernetesClient.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Updated pod label: %s\n", updatedPod.Labels["leader"])
}

func runLeaderElection(lock *resourcelock.LeaseLock, ctx context.Context, id string) {
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				mutex.Unlock()
				setLeaderLabel(true)
				worker()
			},
			OnStoppedLeading: func() {
				log.Warn("We are no longer the leader, terminating...")
				mutex.Lock()
				setLeaderLabel(false)

				os.Exit(0)
			},
			OnNewLeader: func(current_id string) {
				if current_id == id {
					log.Info("We are still the leading!")

					return
				}
				log.Warnf("New leader is %s", current_id)
			},
		},
	})
}
