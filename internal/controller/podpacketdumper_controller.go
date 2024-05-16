package controller
import (
    "context"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    tcpdumpv1alpha1 "github.com/saguanamu/kn-operator-sdk.git/api/v1alpha1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)
// PodPacketDumperReconciler reconciles a PodPacketDumper object
type PodPacketDumperReconciler struct {
        client.Client
        Scheme *runtime.Scheme
}

func (r *PodPacketDumperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // Fetch the PodPacketDumper instance
    podPacketDumper := &tcpdumpv1alpha1.PodPacketDumper{}
    err := r.Get(ctx, req.NamespacedName, podPacketDumper)
    if err != nil {
        if errors.IsNotFound(err) {
            // Requested PodPacketDumper resource not found. Ignoring since object must be deleted.
            return reconcile.Result{}, nil
        }
        // Error reading the object - requeue the request.
        logger.Error(err, "Failed to get PodPacketDumper")
        return reconcile.Result{}, err
    }

    // Check if the Pod already exists, if not, create a new one
    pod := &corev1.Pod{}
    err = r.Get(ctx, client.ObjectKey{Name: "test-pod", Namespace: "default"}, pod)
    if err != nil && errors.IsNotFound(err) {
        // Define a new Pod
        newPod := r.newPodForPacketDumper(podPacketDumper)
        if err := controllerutil.SetControllerReference(podPacketDumper, newPod, r.Scheme); err != nil {
            logger.Error(err, "Failed to set owner reference on Pod")
            return reconcile.Result{}, err
        }

        // Create the Pod
        if err := r.Create(ctx, newPod); err != nil {
            logger.Error(err, "Failed to create Pod")
            return reconcile.Result{}, err
        }
        logger.Info("Pod created", "Name", newPod.Name)
        return reconcile.Result{Requeue: true}, nil
    } else if err != nil {
        logger.Error(err, "Failed to get Pod")
        return reconcile.Result{}, err
    }

    // Pod already exists, just log the event
    logger.Info("Pod already exists", "Name", pod.Name)
    return reconcile.Result{}, nil
}

// newPodForPacketDumper returns a new Pod for the given PodPacketDumper instance
func (r *PodPacketDumperReconciler) newPodForPacketDumper(packetDumper *tcpdumpv1alpha1.PodPacketDumper) *corev1.Pod {
    labels := map[string]string{
        "app": "packet-dumper",
    }

    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
            Labels:    labels,
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  "packet-dumper-container",
                    Image: "hello.io/loopy",
                },
            },
        },
    }

    return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodPacketDumperReconciler) SetupWithManager(mgr ctrl.Manager) error {
        return ctrl.NewControllerManagedBy(mgr).
                For(&tcpdumpv1alpha1.PodPacketDumper{}).
                Complete(r)
} 
