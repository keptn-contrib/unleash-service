# How to manually test unleash-service

1. Install and configure Unleash (out of scope here)
2. Install unleash-service (see [Installation instructions](../README.md#Installation))
3. Create Kubernetes secret for unleash-service to connect to Unleash: `kubectl -n keptn create secret generic unleash --from-literal="UNLEASH_SERVER_URL=http://unleash.unleash-dev/api" --from-literal="UNLEASH_USER=keptn" --from-literal="UNLEASH_TOKEN=keptn"`
4. Create your Keptn project with a stage `production` and a sequence `remediation`: `keptn create project sockshop --shipyard=shipyard.yaml`
5. Create a service: `keptn create service carts --project sockshop`
6. Add remediation.yaml to your project: `keptn add-resource --project sockshop --service carts --stage production --resource remediation.yaml --resourceUri remediation.yaml`
7. Trigger remediation using CLI: `keptn send event -f remediation-triggered.json`
8. Watch :)
