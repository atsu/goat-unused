package k8s

import (
	"log"
	"time"

	osappsv1 "github.com/openshift/api/apps/v1"
	authorizationv1 "github.com/openshift/api/authorization/v1"
	buildv1 "github.com/openshift/api/build/v1"
	osimagev1 "github.com/openshift/api/image/v1"
	networkv1 "github.com/openshift/api/network/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	projectv1 "github.com/openshift/api/project/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	osroutev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	templatev1 "github.com/openshift/api/template/v1"
	userv1 "github.com/openshift/api/user/v1"
	appsv1client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imagev1client "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	routev1client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func init() {
	// The Kubernetes Go client (nested within the OpenShift Go client)
	// automatically registers its types in scheme.Scheme, however the
	// additional OpenShift types must be registered manually.  AddToScheme
	// registers the API group types (e.g. route.openshift.io/v1, Route) only.
	logFatal(osappsv1.AddToScheme(scheme.Scheme))
	logFatal(authorizationv1.AddToScheme(scheme.Scheme))
	logFatal(buildv1.AddToScheme(scheme.Scheme))
	logFatal(osimagev1.AddToScheme(scheme.Scheme))
	logFatal(networkv1.AddToScheme(scheme.Scheme))
	logFatal(oauthv1.AddToScheme(scheme.Scheme))
	logFatal(projectv1.AddToScheme(scheme.Scheme))
	logFatal(quotav1.AddToScheme(scheme.Scheme))
	logFatal(osroutev1.AddToScheme(scheme.Scheme))
	logFatal(securityv1.AddToScheme(scheme.Scheme))
	logFatal(templatev1.AddToScheme(scheme.Scheme))
	logFatal(userv1.AddToScheme(scheme.Scheme))

	// If you need to serialize/deserialize legacy (non-API group) OpenShift
	// types (e.g. v1, Route), these must be additionally registered using
	// AddToSchemeInCoreGroup.
	logFatal(osappsv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(authorizationv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(buildv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(osimagev1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(networkv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(oauthv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(projectv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(quotav1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(osroutev1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(securityv1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(templatev1.AddToSchemeInCoreGroup(scheme.Scheme))
	logFatal(userv1.AddToSchemeInCoreGroup(scheme.Scheme))
}
func logFatal(err interface{}) {
	if err != nil {
		log.Fatalln("FATAL", err)
	}
}

type OpenshiftClient interface {
	K8Client
	IsIngressBackedRoute(route *osroutev1.Route) bool
	ListDeploymentConfigs(namespace string, listOptions metav1.ListOptions) (*osappsv1.DeploymentConfigList, error)
	ListImageStreams(namespace string, listOptions metav1.ListOptions) (*osimagev1.ImageStreamList, error)
	ListRoutes(namespace string, listOptions metav1.ListOptions) (*osroutev1.RouteList, error)
	UpdateDeploymentConfig(namespace string, obj *osappsv1.DeploymentConfig) error
	UpdateImageStream(namespace string, obj *osimagev1.ImageStream) error
	UpdateRoute(namespace string, obj *osroutev1.Route) error
}

var _ OpenshiftClient = &defaultOpenshiftClient{}

type defaultOpenshiftClient struct {
	K8Client
	osApps  appsv1client.AppsV1Interface
	osImg   imagev1client.ImageV1Interface
	osRoute routev1client.RouteV1Interface
}

func NewDefaultOpenshiftController() *defaultOpenshiftClient {
	return &defaultOpenshiftClient{
		K8Client: NewDefaultK8Client(),
	}
}

func (r *defaultOpenshiftClient) Authenticate(rc *rest.Config) error {
	if err := r.K8Client.Authenticate(rc); err != nil {
		return err
	}
	config := r.K8Client.Config()
	r.osImg = imagev1client.NewForConfigOrDie(config)
	r.osRoute = routev1client.NewForConfigOrDie(config)
	r.osApps = appsv1client.NewForConfigOrDie(config)
	return nil
}

func (r *defaultOpenshiftClient) SupportedObject(object runtime.Object) (bool, string) {
	shortName := ""

	switch object.(type) {
	case *osappsv1.DeploymentConfig:
		shortName = "dc"
	case *osimagev1.ImageStream:
		shortName = "is"
	case *osroutev1.Route:
		shortName = "r"
	default:
		return r.K8Client.SupportedObject(object)
	}
	return true, shortName
}

func (r *defaultOpenshiftClient) DeleteObject(namespace string, object runtime.Object) error {
	var err error
	deletePolicy := metav1.DeletePropagationForeground
	options := &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}

	switch t := object.(type) {
	case *osroutev1.Route:
		err = r.osRoute.Routes(namespace).Delete(t.Name, options)
	case *osappsv1.DeploymentConfig:
		err = r.osApps.DeploymentConfigs(namespace).Delete(t.Name, options)
	case *osimagev1.ImageStream:
		err = r.osImg.ImageStreams(namespace).Delete(t.Name, options)
	default:
		err = r.K8Client.DeleteObject(namespace, object)
	}
	return err
}

func (r *defaultOpenshiftClient) DeleteCollection(namespace string, obj runtime.Object, opts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	switch obj.(type) {
	case *osappsv1.DeploymentConfig:
		return r.osApps.DeploymentConfigs(namespace).DeleteCollection(opts, listOpts)
	case *osimagev1.ImageStream:
		return r.osImg.ImageStreams(namespace).DeleteCollection(opts, listOpts)
	case *osroutev1.Route:
		return r.osRoute.Routes(namespace).DeleteCollection(opts, listOpts)
	default:
		return r.K8Client.DeleteCollection(namespace, obj, opts, listOpts)
	}
}

func (r *defaultOpenshiftClient) ListImageStreams(namespace string, listOptions metav1.ListOptions) (*osimagev1.ImageStreamList, error) {
	listOptions.Limit = 100
	list, err := r.osImg.ImageStreams(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "ImageStream", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultOpenshiftClient) UpdateImageStream(namespace string, obj *osimagev1.ImageStream) error {
	old, err := r.osImg.ImageStreams(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.osImg.ImageStreams(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.osImg.ImageStreams(namespace).Create(obj)
	}
	return err
}

func (r *defaultOpenshiftClient) UpdateObject(namespace string, object runtime.Object) error {
	switch t := object.(type) {
	case *osroutev1.Route:
		return r.UpdateRoute(namespace, t)
	case *osappsv1.DeploymentConfig:
		return r.UpdateDeploymentConfig(namespace, t)
	case *osimagev1.ImageStream:
		return r.UpdateImageStream(namespace, t)
	default:
		return r.K8Client.UpdateObject(namespace, object)
	}
}

func (r *defaultOpenshiftClient) ListDeploymentConfigs(namespace string, listOptions metav1.ListOptions) (*osappsv1.DeploymentConfigList, error) {
	listOptions.Limit = 100
	list, err := r.osApps.DeploymentConfigs(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "DeploymentConfig", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultOpenshiftClient) UpdateDeploymentConfig(namespace string, obj *osappsv1.DeploymentConfig) error {
	old, err := r.osApps.DeploymentConfigs(namespace).Get(obj.Name, metav1.GetOptions{})
	if err == nil {
		obj.ResourceVersion = old.ResourceVersion
		_, err = r.osApps.DeploymentConfigs(namespace).Update(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.osApps.DeploymentConfigs(namespace).Create(obj)
	}
	return err
}

func (r *defaultOpenshiftClient) ListRoutes(namespace string, listOptions metav1.ListOptions) (*osroutev1.RouteList, error) {
	listOptions.Limit = 100
	list, err := r.osRoute.Routes(namespace).List(listOptions)
	if err == nil {
		for i := range list.Items {
			list.Items[i].TypeMeta = metav1.TypeMeta{Kind: "Route", APIVersion: "v1"}
		}
	}
	return list, err
}

func (r *defaultOpenshiftClient) UpdateRoute(namespace string, obj *osroutev1.Route) error {
	old, err := r.osRoute.Routes(namespace).Get(obj.Name, metav1.GetOptions{})

	if err == nil {
		log.Println("deleting route", old.Name, old.UID)
		err = r.DeleteObject(namespace, obj)
		if err != nil {
			return err
		}
		log.Println("sleeping 5 seconds after deleting route ", old.Name, old.UID)
		time.Sleep(5 * time.Second)

		_, err = r.osRoute.Routes(namespace).Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				log.Println("route", obj.Name, old.UID, "is deleted")
			} else {
				return err
			}
		} else {
			log.Println("sleeping 20 more seconds, waiting for route delete")
			time.Sleep(20 * time.Second)
		}

		log.Println("recreating route", obj.Name)
		_, err = r.osRoute.Routes(namespace).Create(obj)
	} else if k8serrors.IsNotFound(err) {
		_, err = r.osRoute.Routes(namespace).Create(obj)
	}

	return err
}

func (r *defaultOpenshiftClient) ListObjects(namespace string, options metav1.ListOptions) (*corev1.List, error) {
	list, err := r.K8Client.ListObjects(namespace, options)
	if err != nil {
		return nil, err
	}

	routes, err := r.ListRoutes(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range routes.Items {
		re := runtime.RawExtension{Object: &routes.Items[i]}
		list.Items = append(list.Items, re)
	}

	dcs, err := r.ListDeploymentConfigs(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range dcs.Items {
		re := runtime.RawExtension{Object: &dcs.Items[i]}
		list.Items = append(list.Items, re)
	}

	is, err := r.ListImageStreams(namespace, options)
	if err != nil {
		return nil, err
	}
	for i := range is.Items {
		re := runtime.RawExtension{Object: &is.Items[i]}
		list.Items = append(list.Items, re)
	}
	return list, nil
}

func (r *defaultOpenshiftClient) IsIngressBackedRoute(route *osroutev1.Route) bool {
	if route == nil {
		return false
	}
	for _, r := range route.OwnerReferences {
		if r.Kind == "Ingress" && *r.Controller {
			return true
		}
	}
	return false
}

func (r *defaultOpenshiftClient) DeleteList(namespace string, list *corev1.List) error {
	for _, listItem := range list.Items {
		object := listItem.Object.(runtime.Object)
		err := r.DeleteObject(namespace, object)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *defaultOpenshiftClient) UpdateList(namespace string, list *corev1.List) error {
	for _, listItem := range list.Items {
		object := listItem.Object.(runtime.Object)
		err := r.UpdateObject(namespace, object)
		if err != nil {
			return err
		}
	}
	return nil
}
