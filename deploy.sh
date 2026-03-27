#!/bin/bash

PROJECT_ID="plinehandotcom-hrd"
APP_ENGINE_SA="${PROJECT_ID}@appspot.gserviceaccount.com"
STAGING_BUCKET="gs://staging.${PROJECT_ID}.appspot.com"

gcloud storage buckets add-iam-policy-binding "$STAGING_BUCKET" \
  --member="serviceAccount:${APP_ENGINE_SA}" \
  --role="roles/storage.objectAdmin"

gcloud app deploy --project=plinehandotcom-hrd
