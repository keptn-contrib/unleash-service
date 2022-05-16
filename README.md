# Unleash Service

This service allows to interact with the open source feature toggle system [unleash](https://github.com/unleash). 
Triggered by a Keptn CloudEvent of the type `sh.keptn.event.action.triggered`. After the features specified in the event 
have been toggled, it sends out an `sh.keptn.event.action.finished` event.

Example payload for an action.triggered event:

```
{
  "type": "sh.keptn.event.action.triggered",
  "specversion": "1.0",
  "source": "keptn-user",
  "id": "f2b878d3-03c0-4e8f-bc3f-454bc1b3d79d",
  "time": "2019-06-07T07:02:15.64489Z",
  "contenttype": "application/json",
  "shkeptncontext": "08735340-6f9e-4b32-97ff-3b6c292bc509",
  "data": {
    "action": {
      "name": "FeatureToggle",
      "action": "toggle-feature",
      "description": "toggle a feature",
      "values": {
        "EnableItemCache": "on"
      }
    },
    "problem": {
      "ImpactedEntity": "carts-primary",
      "PID": "93a5-3fas-a09d-8ckf",
      "ProblemDetails": "Pod name",
      "ProblemID": "762",
      "ProblemTitle": "cpu_usage_sockshop_carts",
      "State": "OPEN"
    },
    "project": "sockshop",
    "stage": "staging",
    "service": "carts",
    "labels": {
      "testid": "12345",
      "buildnr": "build17",
      "runby": "JohnDoe"
    }
  }
}
```

## Compatibility Matrix

Please always double check the version of Keptn you are using compared to the version of this service, and follow the compatibility matrix below.


| Keptn Version\*  | [Unleash Service Image](https://hub.docker.com/r/keptncontrib/unleash-service/tags) |
|:----------------:|:----------------------------------------:|
|       0.6.x      | keptncontrib/unleash-service:0.1.0  |
|       0.7.x      | keptncontrib/unleash-service:0.2.0  |
|       0.8.x      | keptncontrib/unleash-service:0.3.0  |
|    0.8.0-0.8.3   | keptncontrib/unleash-service:0.3.1  |
|       0.8.4      | keptncontrib/unleash-service:0.3.2  |

\* This is the Keptn version we aim to be compatible with. Other versions should work too, but there is no guarantee.

**Note**: This integration is currently incompatible with Keptn 0.14.x and newer.

## Installation

The *unleash-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *unleash-service* in your Keptn Kubernetes cluster use the [`helm chart`](chart/Chart.yaml) file, for example:

```console
helm install -n keptn unleash-service chart/
```

This should install the `unleash-service` together with a Keptn `distributor` into the `keptn` namespace, which you can verify using

```console
kubectl -n keptn get deployment unleash-service -o wide
kubectl -n keptn get pods -l run=unleash-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
helm upgrade -n keptn --set image.tag=$VERSION unleash-service chart/
```

### Uninstall

To delete a deployed *unleash-service*, use the file `deploy/*.yaml` files from this repository and delete the Kubernetes resources:

```console
helm uninstall -n keptn unleash-service
```

## Development

Development can be conducted using any GoLang compatible IDE/editor (e.g., Jetbrains GoLand, VSCode with Go plugins).

It is recommended to make use of branches as follows:

* `main`/`master` contains the latest potentially unstable version
* `release-*` contains a stable version of the service (e.g., `release-0.1.0` contains version 0.1.0)
* create a new branch for any changes that you are working on, e.g., `feature/my-cool-stuff` or `bug/overflow`
* once ready, create a pull request from that branch back to the `main`/`master` branch

When writing code, it is recommended to follow the coding style suggested by the [Golang community](https://github.com/golang/go/wiki/CodeReviewComments).
