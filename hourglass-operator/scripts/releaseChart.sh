#!/usr/bin/env bash

bucket_name="eigenlabs-hourglass-dev"
helm_repo_url="https://eigenlabs-hourglass-dev.s3.amazonaws.com/helm"

mkdir chart_releases || true

helm package ./charts/* --destination chart_releases

if aws s3 ls "s3://${bucket_name}/helm/index.yaml" &>/dev/null; then
    echo "Downloading existing index.yaml"
    aws s3 cp "s3://${bucket_name}/helm/index.yaml" ./chart_releases/

    echo "Generating index"
    helm repo index --merge ./chart_releases/index.yaml --url $helm_repo_url ./chart_releases
else
    echo "Generating index for the first time"
    helm repo index --url $helm_repo_url ./chart_releases
fi

aws s3 sync ./chart_releases/ "s3://${bucket_name}/helm"
