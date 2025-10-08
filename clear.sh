#!/bin/bash
set -e

BUCKETS=(
  "dev-static-website-websitedistributionloggingbucke-gnywvopp3cfp"
  "dev-static-website-websitedistributionloggingbucke-i95lggy24ss5"
  "dev-static-website-websitedistributionloggingbucke-ljnc5ai09l3y"
  "dev-static-website-websitedistributionloggingbucke-myadtspuaoj9"
  "dev-static-website-websitedistributionloggingbucke-ov0bdtq3zkxl"
  "dev-static-website-websitedistributionloggingbucke-rtrevzwrpcos"
  "dev-static-website-websitedistributionloggingbucke-snx3j73c0xlv"
  "dev-static-website-websitedistributionloggingbucke-wdjccbkmuuxe"
  "dev-static-website-websitedistributionloggingbucke-wwakp4c8yjlm"
  "dev-static-website-websitedistributionloggingbucke-z2mnimm3uf8z"
)

AWS_PROFILE="default"
AWS_REGION="us-east-2"

for BUCKET in "${BUCKETS[@]}"; do
  echo "🔹 Eliminando bucket: $BUCKET"
  aws s3 rb "s3://$BUCKET" --force --profile $AWS_PROFILE --region $AWS_REGION
  echo "✅ Bucket $BUCKET eliminado."
done

echo "🎉 Todos los buckets eliminados."
