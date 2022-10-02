package main

import (
	"os"
	"context"
	"flag"
	"strconv"
	"github.com/rs/zerolog"
	"github.com/rs/xid"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	webhookName            string
	objectMetaName         string
	serviceName            string
	serviceNamespace       string
	servicePath            string
	namespaceSelectorKey   string
	namespaceSelectorValue string
	certPath               string
	keyPath                string
	createSecret			 		 bool
	log										 zerolog.Logger
) 

func main() {
	flag.StringVar(&webhookName, "webhook-name", "", "Webhook name for MutatingWebhookConfiguration, required")
	flag.StringVar(&objectMetaName, "object-meta-name", "", "ObjectMeta name for MutatingWebhookConfiguration, required")
	flag.StringVar(&serviceName, "service-name", "", "ClientConfig service name, required")
	flag.StringVar(&serviceNamespace, "service-namespace", "", "ClientConfig service namespace, required")
	flag.StringVar(&servicePath, "service-path", "", "ClientConfig service path, required")
	flag.StringVar(&namespaceSelectorKey, "namespace-selector-key", "", "namespaceSelector matchLabels key, required")
	flag.StringVar(&namespaceSelectorValue, "namespace-selector-value", "", "namespaceSelector matchLabels value, required")
	flag.StringVar(&certPath, "cert-path", "/etc/tls/tls.crt", "The path/to/file where the TLS certs are found")
	flag.StringVar(&keyPath, "key-path", "/etc/tls/tls.key", "The path/to/file where the TLS certs are found")
	flag.BoolVar(&createSecret, "create-secret", false, "Create kubernetes secret from certificate and private key data")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log = zerolog.New(os.Stdout).With().Str("correlationID", xid.New().String()).Str("container", "mutatingwebhook-init-container").Str("service", objectMetaName).Timestamp().Logger()
	
	log.Info().Msg("Requested MutatingWebhookConfiguration")
	log.Info().Str("--webhook-name=", webhookName).Send()
	log.Info().Str("--object-meta-name=", objectMetaName).Send()
	log.Info().Str("--service-name=", serviceName).Send()
	log.Info().Str("--service-namespace=", serviceNamespace).Send()
	log.Info().Str("--service-path=", servicePath).Send()
	log.Info().Str("--namespace-selector-key=", namespaceSelectorKey).Send()
	log.Info().Str("--namespace-selector-value=", namespaceSelectorValue).Send()
	log.Info().Str("--cert-path=", certPath).Send()
	log.Info().Str("--key-path=", keyPath).Send()
	log.Info().Str("--create-secret=", strconv.FormatBool(createSecret)).Send()

	// read CA Bundle
	certificatePEM, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read CA Bundle")
	}

	// read CA private key
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read private key")
	}

	// create kubernetes api client
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set go-client")
	}

	// create mutatingwebhookconfiguration resource definition
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
			AdmissionReviewVersions: []string{"v1beta1"},
			SideEffects:             &none,
		}},
	}

	// fetch the list of existing mutatingwebhookconfigurations
	mutatingWebhookList, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("fail to list mutatingwebhooks")
	}
	// if the requested webhook already exists, get the ResourceVersion so it can be updated
	mutateconfig.ObjectMeta.ResourceVersion = webhookExists(mutatingWebhookList, objectMetaName)

	// create, or update if already exists
	if mutateconfig.ObjectMeta.ResourceVersion != "" {
		if _, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), mutateconfig, metav1.UpdateOptions{}); err != nil {
			log.Fatal().Err(err).Msg("failed to update mutatingwebhook")
		}
		log.Info().Msg("Success: updated mutatingwebhookconfiguration")
	} else {
		if _, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), mutateconfig, metav1.CreateOptions{}); err != nil {
			log.Fatal().Err(err).Msg("failed to create mutatingwebhook")
		}
		log.Info().Msg("Success: created mutatingwebhookconfiguration")
	}

	// create kubernetes Secret with certificate and private key data, if requested
	if createSecret {
		log.Info().Msg("create kubernetes secret with certificate and private key data")
		secretData := map[string][]byte{
			"tls.crt": []byte(certificatePEM),
			"tls.key": []byte(privateKey),
		}
		secretName := objectMetaName + "-certificate"
		certificateSecret := coreV1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: serviceNamespace,
			},
			Data: secretData,
		}
	
		// fetch the list of existing secrets in the namespace
		NSSecretList, err := kubeClient.CoreV1().Secrets(serviceNamespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create secret")
		}
		// if the requested secret already exists, get the ResourceVersion so it can be updated
		certificateSecret.ObjectMeta.ResourceVersion = secretExists(NSSecretList, secretName)
	
		// create, or update if already exists
		if certificateSecret.ObjectMeta.ResourceVersion != "" {
			if _, err = kubeClient.CoreV1().Secrets(serviceNamespace).Update(context.Background(), &certificateSecret, metav1.UpdateOptions{}); err != nil {
				log.Fatal().Err(err).Msg("failed to update secret")
			}
			log.Info().Msg("Success: updated certificate secret")
		} else {
			if _, err = kubeClient.CoreV1().Secrets(serviceNamespace).Create(context.Background(), &certificateSecret, metav1.CreateOptions{}); err != nil {
				log.Fatal().Err(err).Msg("failed to create secret")
			}
			log.Info().Msg("Success: created certificate secret")
		}
	}
}

// search list of existing mutatingwebhookconfnigurations for match
func webhookExists(webhookList *admissionregistrationv1.MutatingWebhookConfigurationList, objectMetaName string) string {
	log.Info().Msg("checking if mutatingwebhookconfiguration already exists")
	for i := range webhookList.Items {
		log.Trace().Str("searching: ", webhookList.Items[i].Webhooks[0].Name)
		if webhookList.Items[i].ObjectMeta.Name == objectMetaName {
			log.Info().Msg("found, update with ResourceVersion")
			return webhookList.Items[i].ObjectMeta.ResourceVersion
		}
	}
	log.Info().Msg("not found, create new mutatingwebhookconfiguration")
	return ""
}

// search list of existing namespace secrets for match
func secretExists(currentNSSecrets *coreV1.SecretList, secretName string) string {
	log.Info().Str("checking if secret already exists: ", secretName)
	for i := range currentNSSecrets.Items {
		log.Trace().Str("searching:", currentNSSecrets.Items[i].ObjectMeta.Name)
		if currentNSSecrets.Items[i].ObjectMeta.Name == secretName {
			log.Info().Msg("found, update with ResourceVersion")
			return currentNSSecrets.Items[i].ObjectMeta.ResourceVersion
		}
	}
	log.Info().Msg("not found, create new certificate secret")
	return ""
}