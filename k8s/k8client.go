package k8s

import (
	"fmt"
	"log"
	"time"

	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta1api "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8Client interface {
	Authenticate(rc *rest.Config) error
	Config() *rest.Config
	DeleteCollection(namespace string, obj runtime.Object, opts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	DeleteList(namespace string, list *corev1.List) error
	DeleteObject(namespace string, object runtime.Object) error
	GetClusterRole(name string, options metav1.GetOptions) (*v1.ClusterRole, error)
	GetClusterRoleBinding(name string, options metav1.GetOptions) (*v1.ClusterRoleBinding, error)
	GetConfigMap(namespace, name string, options metav1.GetOptions) (*corev1.ConfigMap, error)
	GetDeployment(namespace, name string, options metav1.GetOptions) (*appsv1.Deployment, error)
	GetPersistentVolumeClaim(namespace, name string, options metav1.GetOptions) (*corev1.PersistentVolumeClaim, error)
	GetPod(namespace, name string, options metav1.GetOptions) (*corev1.Pod, error)
	GetPodLogs(namespace, name string, options *corev1.PodLogOptions) ([]byte, error)
	GetRole(namespace, name string, options metav1.GetOptions) (*v1.Role, error)
	GetRoleBinding(namespace, name string, options metav1.GetOptions) (*v1.RoleBinding, error)
	ListClusterRoleBindings(options metav1.ListOptions) (*v1.ClusterRoleBindingList, error)
	ListClusterRoles(options metav1.ListOptions) (*v1.ClusterRoleList, error)
	ListCronJobs(namespace string, options metav1.ListOptions) (*v1beta1.CronJobList, error)
	ListConfigMaps(namespace string, listOptions metav1.ListOptions) (*corev1.ConfigMapList, error)
	ListDeployments(namespace string, listOptions metav1.ListOptions) (*appsv1.DeploymentList, error)
	ListIngresses(namespace string, listOptions metav1.ListOptions) (*v1beta1api.IngressList, error)
	ListJobs(namespace string, listOptions metav1.ListOptions) (*batchv1.JobList, error)
	ListNamespaces(options metav1.ListOptions) (*corev1.NamespaceList, error)
	ListObjects(namespace string, listOptions metav1.ListOptions) (*corev1.List, error)
	ListPersistentVolumeClaims(namespace string, listOptions metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error)
	ListPods(namespace string, listOptions metav1.ListOptions) (*corev1.PodList, error)
	ListRoles(namespace string, options metav1.ListOptions) (*v1.RoleList, error)
	ListRoleBindings(namespace string, options metav1.ListOptions) (*v1.RoleBindingList, error)
	ListSecrets(namespace string, listOptions metav1.ListOptions) (*corev1.SecretList, error)
	ListServices(namespace string, listOptions metav1.ListOptions) (*corev1.ServiceList, error)
	ListStatefulSets(namespace string, listOptions metav1.ListOptions) (*appsv1.StatefulSetList, error)
	LogServerInfo() (*ServerInfo, error)
	PatchDeployment(namespace, name string, patch []byte) (*appsv1.Deployment, error)
	SupportedObject(object runtime.Object) (bool, string)
	UpdateClusterRole(obj *v1.ClusterRole) error
	UpdateClusterRoleBinding(obj *v1.ClusterRoleBinding) error
	UpdateConfigMap(namespace string, obj *corev1.ConfigMap) error
	UpdateDeployment(namespace string, obj *appsv1.Deployment) error
	UpdateIngress(namespace string, obj *v1beta1api.Ingress) error
	UpdateCronJob(namespace string, obj *v1beta1.CronJob) error
	UpdateJob(namespace string, obj *batchv1.Job) error
	UpdateList(namespace string, list *corev1.List) error
	UpdateObject(namespace string, object runtime.Object) error
	UpdatePersistentVolumeClaim(namespace string, obj *corev1.PersistentVolumeClaim) error
	UpdatePod(namespace string, obj *corev1.Pod) error
	UpdateRole(namespace string, obj *v1.Role) error
	UpdateRoleBinding(namespace string, obj *v1.RoleBinding) error
	UpdateSecret(namespace string, obj *corev1.Secret) error
	UpdateService(namespace string, obj *corev1.Service) error
	UpdateStatefulSet(namespace string, obj *appsv1.StatefulSet) error
	WatchEvents(namespace string, listOptions metav1.ListOptions) (watch.Interface, error)
}

var _ K8Client = &defaultK8Client{}

type defaultK8Client struct {
	rc        *rest.Config
	k8Clients *kubernetes.Clientset
}

func (r *defaultK8Client) Config() *rest.Config {
	return r.rc
}

func NewDefaultK8Client() *defaultK8Client {
	return &defaultK8Client{}
}

func (r *defaultK8Client) SupportedObject(object runtime.Object) (bool, string) {
	shortName := ""

	switch object.(type) {
	case *corev1.ConfigMap:
		shortName = "cm"
	case *appsv1.Deployment:
		shortName = "deploy"
	case *v1beta1api.Ingress:
		shortName = "ing"
	case *batchv1.Job:
		shortName = "job"
	case *v1beta1.CronJob:
		shortName = "cronjob"
	case *corev1.Secret:
		shortName = "secret"
	case *corev1.Service:
		shortName = "svc"
	case *corev1.PersistentVolumeClaim:
		shortName = "pvc"
	case *appsv1.StatefulSet:
		shortName = "ss"
	case *v1.Role:
		shortName = "role"
	case *v1.RoleBinding:
		shortName = "rolebinding"
	case *v1.ClusterRole:
		shortName = "clusterrole"
	case *v1.ClusterRoleBinding:
		shortName = "clusterrolebinding"
	default:
		return false, "unknown"
	}
	return true, shortName
}

func (r *defaultK8Client) Authenticate(rc *rest.Config) (err error) {
	if rc == nil { // if no rest config passed in, we try for in cluster
		rc, err = rest.InClusterConfig()
		if err != nil {
			return
		}
	}
	r.rc = rc
	r.k8Clients = kubernetes.NewForConfigOrDie(rc)
	return nil
}

func (r *defaultK8Client) ListNamespaces(options metav1.ListOptions) (*corev1.NamespaceList, error) {
	return r.k8Clients.CoreV1().Namespaces().List(options)
}

func (r *defaultK8Client) ListRoles(namespace string, options metav1.ListOptions) (*v1.RoleList, error) {
	return r.k8Clients.RbacV1().Roles(namespace).List(options)
}

func (r *defaultK8Client) ListRoleBindings(namespace string, options metav1.ListOptions) (*v1.RoleBindingList, error) {
	return r.k8Clients.RbacV1().RoleBindings(namespace).List(options)
}

func (r *defaultK8Client) ListClusterRoles(options metav1.ListOptions) (*v1.ClusterRoleList, error) {
	return r.k8Clients.RbacV1().ClusterRoles().List(options)
}

func (r *defaultK8Client) ListClusterRoleBindings(options metav1.ListOptions) (*v1.ClusterRoleBindingList, error) {
	return r.k8Clients.RbacV1().ClusterRoleBindings().List(options)
}

func (r *defaultK8Client) ListCronJobs(namespace string, options metav1.ListOptions) (*v1beta1.CronJobList, error) {
	return r.k8Clients.BatchV1beta1().CronJobs(namespace).List(options)
}

type ServerInfo struct {
	Document *openapi_v2.Document
	Info     *version.Info
}

func (r *defaultK8Client) LogServerInfo() (*ServerInfo, error) {
	document, err := r.k8Clients.DiscoveryClient.OpenAPISchema()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve document: %v", err)
	}

	info, err := r.k8Clients.DiscoveryClient.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server info: %v", err)
	}
	si := &ServerInfo{
		Document: document,
		Info:     info,
	}
	return si, nil
}

func (r *defaultK8Client) DeleteList(namespace string, list *corev1.List) error {
	for _, listItem := range list.Items {
		object := listItem.Object.(runtime.Object)
		err := r.DeleteObject(namespace, object)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *defaultK8Client) UpdateList(namespace string, list *corev1.List) error {
	for _, listItem := range list.Items {
		object := listItem.Object.(runtime.Object)
		err := r.UpdateObject(namespace, object)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *defaultK8Client) DeleteObject(namespace string, object runtime.Object) error {
	var err error
	deletePolicy := metav1.DeletePropagationForeground
	options := &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}

	switch t := object.(type) {
	case *v1.Role:
		err = r.k8Clients.RbacV1().Roles(namespace).Delete(t.Name, options)
	case *v1.RoleBinding:
		err = r.k8Clients.RbacV1().RoleBindings(namespace).Delete(t.Name, options)
	case *corev1.ConfigMap:
		err = r.k8Clients.CoreV1().ConfigMaps(namespace).Delete(t.Name, options)
	case *corev1.Secret:
		err = r.k8Clients.CoreV1().Secrets(namespace).Delete(t.Name, options)
	case *corev1.Service:
		err = r.k8Clients.CoreV1().Services(namespace).Delete(t.Name, options)
	case *corev1.Pod:
		err = r.k8Clients.CoreV1().Pods(namespace).Delete(t.Name, options)
	case *corev1.PersistentVolumeClaim:
		err = r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).Delete(t.Name, options)
	case *appsv1.Deployment:
		err = r.k8Clients.AppsV1().Deployments(namespace).Delete(t.Name, options)
	case *v1beta1api.Ingress:
		err = r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).Delete(t.Name, options)
	case *batchv1.Job:
		err = r.k8Clients.BatchV1().Jobs(namespace).Delete(t.Name, options)
	case *v1beta1.CronJob:
		err = r.k8Clients.BatchV1beta1().CronJobs(namespace).Delete(t.Name, options)
	case *appsv1.StatefulSet:
		err = r.k8Clients.AppsV1().StatefulSets(namespace).Delete(t.Name, options)
	default:
		err = fmt.Errorf("unsupported object type: %T (%v)", object, object)
	}
	return err
}

func (r *defaultK8Client) UpdateObject(namespace string, object runtime.Object) error {
	switch t := object.(type) {
	case *v1.Role:
		return r.UpdateRole(namespace, t)
	case *v1.RoleBinding:
		return r.UpdateRoleBinding(namespace, t)
	case *corev1.ConfigMap:
		return r.UpdateConfigMap(namespace, t)
	case *corev1.Secret:
		return r.UpdateSecret(namespace, t)
	case *corev1.Service:
		return r.UpdateService(namespace, t)
	case *corev1.Pod:
		return r.UpdatePod(namespace, t)
	case *corev1.PersistentVolumeClaim:
		return r.UpdatePersistentVolumeClaim(namespace, t)
	case *appsv1.Deployment:
		return r.UpdateDeployment(namespace, t)
	case *v1beta1api.Ingress:
		return r.UpdateIngress(namespace, t)
	case *batchv1.Job:
		return r.UpdateJob(namespace, t)
	case *v1beta1.CronJob:
		return r.UpdateCronJob(namespace, t)
	case *appsv1.StatefulSet:
		return r.UpdateStatefulSet(namespace, t)
	default:
		return fmt.Errorf("unsupported object type: %T (%v)", object, object)
	}
}

func (r *defaultK8Client) ListSecrets(namespace string, listOptions metav1.ListOptions) (*corev1.SecretList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.CoreV1().Secrets(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateSecret(namespace string, obj *corev1.Secret) error {
	old, err := r.k8Clients.CoreV1().Secrets(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.CoreV1().Secrets(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.CoreV1().Secrets(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListServices(namespace string, listOptions metav1.ListOptions) (*corev1.ServiceList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.CoreV1().Services(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateService(namespace string, obj *corev1.Service) error {
	old, err := r.k8Clients.CoreV1().Services(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		obj.Spec.ClusterIP = old.Spec.ClusterIP
		_, err = r.k8Clients.CoreV1().Services(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.CoreV1().Services(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListConfigMaps(namespace string, listOptions metav1.ListOptions) (*corev1.ConfigMapList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.CoreV1().ConfigMaps(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) GetConfigMap(namespace, name string, options metav1.GetOptions) (*corev1.ConfigMap, error) {
	configMap, err := r.k8Clients.CoreV1().ConfigMaps(namespace).Get(name, options)
	if err == nil {
		configMap.TypeMeta = metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}
	}
	return configMap, err
}

func (r *defaultK8Client) UpdateConfigMap(namespace string, obj *corev1.ConfigMap) error {
	old, err := r.GetConfigMap(namespace, obj.GetName(), metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.CoreV1().ConfigMaps(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.CoreV1().ConfigMaps(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListStatefulSets(namespace string, listOptions metav1.ListOptions) (*appsv1.StatefulSetList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.AppsV1().StatefulSets(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateStatefulSet(namespace string, obj *appsv1.StatefulSet) error {
	old, err := r.k8Clients.AppsV1().StatefulSets(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.AppsV1().StatefulSets(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.AppsV1().StatefulSets(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListJobs(namespace string, listOptions metav1.ListOptions) (*batchv1.JobList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.BatchV1().Jobs(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Job", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateCronJob(namespace string, obj *v1beta1.CronJob) error {
	old, err := r.k8Clients.BatchV1beta1().CronJobs(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.BatchV1beta1().CronJobs(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.BatchV1beta1().CronJobs(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) UpdateJob(namespace string, obj *batchv1.Job) error {
	old, err := r.k8Clients.BatchV1().Jobs(namespace).Get(obj.Name, metav1.GetOptions{})
	recreate := false

	if err == nil {
		if old.Status.CompletionTime != nil && old.Status.Failed == 1 {
			recreate = true
			log.Println("deleting failed job", old.Name, old.UID)
			err = r.DeleteObject(namespace, obj)
			if err != nil {
				return err
			}
			log.Println("sleeping 10 seconds after deleting job ", old.Name, old.UID)
			time.Sleep(10 * time.Second)
		}
	}
	if k8serrors.IsNotFound(err) || recreate {
		if recreate {
			log.Println("recreating job", obj)
		}
		_, err = r.k8Clients.BatchV1().Jobs(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) GetDeployment(namespace, name string, options metav1.GetOptions) (*appsv1.Deployment, error) {
	deployment, err := r.k8Clients.AppsV1().Deployments(namespace).Get(name, options)
	if err == nil {
		deployment.TypeMeta = metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}
	}
	return deployment, err
}

func (r *defaultK8Client) PatchDeployment(namespace, name string, patch []byte) (*appsv1.Deployment, error) {
	return r.k8Clients.AppsV1().Deployments(namespace).Patch(name, types.StrategicMergePatchType, patch)
}

func (r *defaultK8Client) ListDeployments(namespace string, listOptions metav1.ListOptions) (*appsv1.DeploymentList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.AppsV1().Deployments(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateDeployment(namespace string, obj *appsv1.Deployment) error {
	old, err := r.k8Clients.AppsV1().Deployments(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.AppsV1().Deployments(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.AppsV1().Deployments(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListIngresses(namespace string, listOptions metav1.ListOptions) (*v1beta1api.IngressList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Ingress", APIVersion: "v1beta1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdateIngress(namespace string, obj *v1beta1api.Ingress) error {
	old, err := r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) ListObjects(namespace string, options metav1.ListOptions) (*corev1.List, error) {
	list := &corev1.List{TypeMeta: metav1.TypeMeta{Kind: "List", APIVersion: "v1"}}
	cms, err := r.ListConfigMaps(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range cms.Items {
		re := runtime.RawExtension{Object: &cms.Items[i]}
		list.Items = append(list.Items, re)
	}

	secrets, err := r.ListSecrets(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range secrets.Items {
		re := runtime.RawExtension{Object: &secrets.Items[i]}
		list.Items = append(list.Items, re)
	}

	services, err := r.ListServices(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range services.Items {
		re := runtime.RawExtension{Object: &services.Items[i]}
		list.Items = append(list.Items, re)
	}

	pods, err := r.ListPods(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		re := runtime.RawExtension{Object: &pods.Items[i]}
		list.Items = append(list.Items, re)
	}

	pvcs, err := r.ListPersistentVolumeClaims(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range pvcs.Items {
		re := runtime.RawExtension{Object: &pvcs.Items[i]}
		list.Items = append(list.Items, re)
	}

	ds, err := r.ListDeployments(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range ds.Items {
		re := runtime.RawExtension{Object: &ds.Items[i]}
		list.Items = append(list.Items, re)
	}

	ings, err := r.ListIngresses(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range ings.Items {
		re := runtime.RawExtension{Object: &ings.Items[i]}
		list.Items = append(list.Items, re)
	}

	js, err := r.ListJobs(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range js.Items {
		re := runtime.RawExtension{Object: &js.Items[i]}
		list.Items = append(list.Items, re)
	}

	ss, err := r.ListStatefulSets(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range ss.Items {
		re := runtime.RawExtension{Object: &ss.Items[i]}
		list.Items = append(list.Items, re)
	}
	return list, nil
}

func (r *defaultK8Client) WatchEvents(namespace string, options metav1.ListOptions) (watch.Interface, error) {
	return r.k8Clients.CoreV1().Events(namespace).Watch(options)
}

func (r *defaultK8Client) GetPod(namespace, name string, options metav1.GetOptions) (*corev1.Pod, error) {
	pod, err := r.k8Clients.CoreV1().Pods(namespace).Get(name, options)
	if err == nil {
		pod.TypeMeta = metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}
	}
	return pod, err
}

func (r *defaultK8Client) ListPods(namespace string, listOptions metav1.ListOptions) (*corev1.PodList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.CoreV1().Pods(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) GetPodLogs(namespace, name string, options *corev1.PodLogOptions) ([]byte, error) {
	return r.k8Clients.CoreV1().Pods(namespace).GetLogs(name, options).DoRaw()
}

func (r *defaultK8Client) UpdatePod(namespace string, obj *corev1.Pod) error {
	old, err := r.GetPod(namespace, obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.CoreV1().Pods(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.CoreV1().Pods(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) DeleteCollection(namespace string, obj runtime.Object, opts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	switch obj.(type) {
	case *corev1.ConfigMap:
		return r.k8Clients.CoreV1().ConfigMaps(namespace).DeleteCollection(opts, listOpts)
	case *appsv1.Deployment:
		return r.k8Clients.AppsV1().Deployments(namespace).DeleteCollection(opts, listOpts)
	case *v1beta1api.Ingress:
		return r.k8Clients.ExtensionsV1beta1().Ingresses(namespace).DeleteCollection(opts, listOpts)
	case *batchv1.Job:
		return r.k8Clients.BatchV1().Jobs(namespace).DeleteCollection(opts, listOpts)
	case *corev1.Secret:
		return r.k8Clients.CoreV1().Secrets(namespace).DeleteCollection(opts, listOpts)
	case *corev1.PersistentVolumeClaim:
		return r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(opts, listOpts)
	case *v1.Role:
		return r.k8Clients.RbacV1().Roles(namespace).DeleteCollection(opts, listOpts)
	case *v1.RoleBinding:
		return r.k8Clients.RbacV1().RoleBindings(namespace).DeleteCollection(opts, listOpts)
	case *v1.ClusterRole:
		return r.k8Clients.RbacV1().ClusterRoles().DeleteCollection(opts, listOpts)
	case *v1.ClusterRoleBinding:
		return r.k8Clients.RbacV1().ClusterRoleBindings().DeleteCollection(opts, listOpts)
	default:
		return fmt.Errorf("object type %T does not support DeleteCollection", obj)
	}
}

func (r *defaultK8Client) GetRole(namespace, name string, options metav1.GetOptions) (*v1.Role, error) {
	return r.k8Clients.RbacV1().Roles(namespace).Get(name, options)
}

func (r *defaultK8Client) GetRoleBinding(namespace, name string, options metav1.GetOptions) (*v1.RoleBinding, error) {
	return r.k8Clients.RbacV1().RoleBindings(namespace).Get(name, options)
}

func (r *defaultK8Client) GetClusterRole(name string, options metav1.GetOptions) (*v1.ClusterRole, error) {
	return r.k8Clients.RbacV1().ClusterRoles().Get(name, options)
}

func (r *defaultK8Client) GetClusterRoleBinding(name string, options metav1.GetOptions) (*v1.ClusterRoleBinding, error) {
	return r.k8Clients.RbacV1().ClusterRoleBindings().Get(name, options)
}

func (r *defaultK8Client) UpdateRole(namespace string, obj *v1.Role) error {
	old, err := r.k8Clients.RbacV1().Roles(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.RbacV1().Roles(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.RbacV1().Roles(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) UpdateRoleBinding(namespace string, obj *v1.RoleBinding) error {
	old, err := r.k8Clients.RbacV1().RoleBindings(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.RbacV1().RoleBindings(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.RbacV1().RoleBindings(namespace).Create(obj)
	}
	return err
}

func (r *defaultK8Client) UpdateClusterRole(obj *v1.ClusterRole) error {
	old, err := r.k8Clients.RbacV1().ClusterRoles().Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.RbacV1().ClusterRoles().Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.RbacV1().ClusterRoles().Create(obj)
	}
	return err
}

func (r *defaultK8Client) UpdateClusterRoleBinding(obj *v1.ClusterRoleBinding) error {
	old, err := r.k8Clients.RbacV1().ClusterRoleBindings().Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.RbacV1().ClusterRoleBindings().Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.RbacV1().ClusterRoleBindings().Create(obj)
	}
	return err
}

func (r *defaultK8Client) GetPersistentVolumeClaim(namespace, name string, options metav1.GetOptions) (*corev1.PersistentVolumeClaim, error) {
	return r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).Get(name, options)
}

func (r *defaultK8Client) ListPersistentVolumeClaims(namespace string, listOptions metav1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	listOptions.Limit = 100
	list, err := r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultK8Client) UpdatePersistentVolumeClaim(namespace string, obj *corev1.PersistentVolumeClaim) error {
	old, err := r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.k8Clients.CoreV1().PersistentVolumeClaims(namespace).Create(obj)
	}
	return err
}
