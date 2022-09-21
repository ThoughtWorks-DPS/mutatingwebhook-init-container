package main

import (
	// "bytes"
	// "crypto/rand"
	// "crypto/rsa"
	// "crypto/x509"
	// "crypto/x509/pkix"
	// "encoding/pem"
	// "fmt"

	"log"

	// "strings"
	// "path"
	// "math/big"
	"os"
	// "time"
	"context"
	"flag"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	webhookName            string
	objectMetaName         string
	serviceName            string
	serviceNamespace       string
	servicePath            string
	caBundle               string
	namespaceSelectorKey   string
	namespaceSelectorValue string
	certPath               string
)

func main() {
	// var certificatePEM *bytes.Buffer
	reviewVersion := []string{"v1beta1"}
	flag.StringVar(&webhookName, "webhook-name", "", "Webhook name for MutatingWebhookConfiguration, required")
	flag.StringVar(&objectMetaName, "object-meta-name", "sidecarinjector.twdps.io", "ObjectMeta name for MutatingWebhookConfiguration, default is sidecarinjector.twdps.io")
	flag.StringVar(&serviceName, "service-name", "", "ClientConfig service name, required")
	flag.StringVar(&serviceNamespace, "service-namespace", "", "ClientConfig service namespace, required")
	flag.StringVar(&servicePath, "service-path", "", "ClientConfig service path, required")
	flag.StringVar(&caBundle, "ca-bundle", "", "CA Bundle, required")
	flag.StringVar(&namespaceSelectorKey, "namespace-selector-key", "", "namespaceSelector matchLabels key, required")
	flag.StringVar(&namespaceSelectorValue, "namespace-selector-value", "", "namespaceSelector matchLabels value, required")
	flag.StringVar(&certPath, "cert-path", "/etc/tls/tls.crt", "The path/to/file where the TLS certs are found")
	flag.Parse()

	log.Println("MutatingWebhookConfiguration deployed with the following information:")
	log.Printf("webhook-name: %s", webhookName)
	log.Printf("object-meta-name: %s", objectMetaName)
	log.Printf("service-name: %s", serviceName)
	log.Printf("service-namespace: %s", serviceNamespace)
	log.Printf("service-path: %s", servicePath)
	log.Printf("ca-bundle: %s", caBundle)
	log.Printf("namespace-selector-key: %s", namespaceSelectorKey)
	log.Printf("namespace-selector-value: %s", namespaceSelectorValue)
	log.Printf("cert-path: %s", certPath)

	// read CA Bundle
	certificatePEM, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("failed to read CA Bundle: %s", err)
	}
	// create kubernetes api client
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to set go -client: %s", err)
	}

	// create mutatingwebhookconfiguration resource request
	fail := admissionregistrationv1.Fail
	none := admissionregistrationv1.SideEffectClassNone
	mutateconfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: objectMetaName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name: webhookName,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: certificatePEM,
				Service: &admissionregistrationv1.ServiceReference{
					Name:      serviceName,
					Namespace: serviceNamespace,
					Path:      &servicePath,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
				admissionregistrationv1.Create},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
				},
			}},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					namespaceSelectorKey: namespaceSelectorValue,
				},
			},
			FailurePolicy:           &fail,
			AdmissionReviewVersions: reviewVersion,
			SideEffects:             &none,
		}},
	}

	listM, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("err listing mutating configs - %v", err)
	}

	log.Printf("%s", listM)

	// if _, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), mutateconfig, metav1.CreateOptions{}); err != nil {
	// 	panic(err)
	// }
	// TODO - Remove when rdy to create
	log.Printf("%s", mutateconfig)
}
