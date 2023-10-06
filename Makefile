# https://cloud.google.com/build/docs/configuring-notifications/configure-http
.EXPORT_ALL_VARIABLES:

WEBHOOK ?= cloudbuild2teams
TRIGGER ?= diva-notifier
PAYLOAD ?= cloudbuild2teams.json
NOTIFIER ?= cloudbuild2teams.yaml
BUCKET ?= diva-notifier
PROJECT ?= diva-production-383303


deploy-webhook:
	gcloud run deploy $(WEBHOOK) --source .

config:
	gsutil cp $(PAYLOAD) gs://$(BUCKET)/$(PAYLOAD)
	gsutil cp $(NOTIFIER) gs://$(BUCKET)/$(NOTIFIER)

deploy-notifier:
	gcloud run deploy $(TRIGGER) \
		--image=us-east1-docker.pkg.dev/gcb-release/cloud-build-notifiers/http:latest \
		--update-env-vars=CONFIG_PATH=gs://$(BUCKET)/$(NOTIFIER),PROJECT_ID=$(PROJECT)
