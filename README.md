# Dynamic Admission Control

This repository is a template for a blogpost about Dynamic Admission Control in Kubernetes.

The webhook contains a mutating and a validating component:
- Mutating: We add a label with a timestamp to a deployment
- Validating: We check if `RunAsRoot` is set in the security context, if it is missing the deployment is rejected

## Building and Running on a kind cluster

Prerequisites:
- CA and certificate / key pair for the webhook server and are created as a k8s secret
- a running [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) cluster


### Building the webhook server image

```bash
make build
```

### Make the image available on the kind cluster

```bash
make pushimage
```

### Creating needed secret

Prerequisites:
- [cfssl](https://github.com/cloudflare/cfssl)

```bash
cd certs
cfssl selfsign inovex-webhook.default.svc csr.json | cfssljson -bare selfsigned

kubectl create secret tls --key selfsigned-key.pem  --cert selfsigned.pem inovex-webhook-certs

**TODO**

The output from the following command needs to be added to `deployment.yml` under $CA_BUNDLE:

echo $(cat selfsigned.pem | base64 | tr -d '\n')
```

### Deployment

After the successful build and push of the image to your kind cluster and the creation of the needed secrets you can deploy the needed components:

- `ValidatingWebhookConfiguration`
- `MutatingWebhookConfiguration`
- Webhook Server Deployment
- Webhook Server Service

This can be done using:
```bash
make deploy
```

### Test the webhook

```bash
kubectl apply -f test_deployment.yml
```

or run:

```bash
make test
```

## Appendix

This code is not production ready. It was written for learning and demonstration purposes.


## License

[MIT](https://choosealicense.com/licenses/mit/)