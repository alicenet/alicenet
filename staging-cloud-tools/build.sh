if [ $# -eq 0 ]; then
    echo './build.sh [CHAIN_ID]'
    exit 1
fi

CHAIN_ID=$1

# Clean storage
gsutil rm gs://chain-$CHAIN_ID-alicenet-builds/**
PROJECT_NAME=$(gcloud projects list | tail -n1 | xargs | cut -d' ' -f1) &&
gsutil rm gs://"$PROJECT_NAME"_cloudbuild/**

BUILD_DIR=".tmp_build"
if [ -d "$BUILD_DIR" ]; then
    rm -rf $BUILD_DIR
fi

mkdir $BUILD_DIR &&
cd $BUILD_DIR &&

git clone git@github.com:alicenet/alicenet.git

cp ../cloudbuild.yaml ./ &&
sed -i -e 's/alicenet-builds/chain-'"$CHAIN_ID"'-alicenet-builds/' cloudbuild.yaml

gcloud builds submit --config ./cloudbuild.yaml ./
