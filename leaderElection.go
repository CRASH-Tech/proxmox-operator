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

func runLeaderElection(lock *resourcelock.LeaseLock, ctx context.Context, id string) {
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				worker()
			},
			OnStoppedLeading: func() {
				log.Warn("We are no longer the leader, terminating...")

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
