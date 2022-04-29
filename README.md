# okta-operator

The [Okta](https://www.okta.com) operator creates Okta applications for ingress resources. The okta operator supports
two annotations:

* `okta.jaconi.io/application` to provide the label for the Okta application created by this operator
* `okta.jaconi.io/trusted-origin` to provide a trusted origin that should be added to Okta.

## Example

Given this ingress resource

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  annotations:
    okta.jaconi.io/application: my-app
    okta.jaconi.io/trusted-origin: my-app.example.com
spec: # [...]
```

the operator will
* create an Okta application with the label `my-app`,
* create a Kubernetes secret `okta-client` containing the applications client ID and secret,
* and add `my-app.example.com` as a trusted origin.

If a group ID is set (see next section), the created app will be added to that group.

## Configuration

To configure the Okta API client, see [https://github.com/okta/okta-sdk-golang#configuration-reference](https://github.com/okta/okta-sdk-golang#configuration-reference).

The simplest solution is to provide two environment variables: `OKTA_CLIENT_TOKEN` and `OKTA_CLIENT_ORGURL`.

The operator itself does not require any configuration. Optionally a group ID can be provided with the `--group-id`
flag. If set, created applications will be added to the group identified by the provided group ID.
