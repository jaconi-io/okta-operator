# okta-operator

The [Okta](https://www.okta.com) operator creates Okta applications for OktaClient Custom Resources. The okta operator supports
the following Custom Resource:

```yaml
apiVersion: okta.jaconi.io/v1alpha1
kind: OktaClient
```

## Example

Given the following

```yaml
apiVersion: okta.jaconi.io/v1alpha1
kind: OktaClient
metadata:
  name: okta-client
spec:
  name: my-app
  clientUri: https://my-app.example.com
  redirectUris:
    - https://my-app.example.com/oauth2/callback
    - https://my-app.example.de/oauth2/callback
  postLogoutRedirectUris:
    - https://my-app.example.com/index.html
    - https://my-app.example.de/index.html
  trustedOrigins:
    - https://my-app.example.com
    - https://my-app.example.de
  groupId: abcdfgh
```

the operator will
* create an Okta application with the label `my-app`,
* create a Kubernetes secret `okta-client` containing the application's client ID and secret,
* and add `my-app.example.com` as well as `my-app.example.de` as a trusted origin.

The created app will be added to the group with the ID `abcdfgh`.

## Configuration

To configure the Okta API client, see [https://github.com/okta/okta-sdk-golang#configuration-reference](https://github.com/okta/okta-sdk-golang#configuration-reference).

The simplest solution is to provide two environment variables: `OKTA_CLIENT_TOKEN` and `OKTA_CLIENT_ORGURL`.

This can be done by providing a ConfigMap and a Secret named `okta` within the Operator's namespace:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: okta
data:
  OKTA_CLIENT_TOKEN: c2VjcmV0
type: Opaque
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: okta
data:
  OKTA_CLIENT_ORGURL: "https://example.oktapreview.com"
```