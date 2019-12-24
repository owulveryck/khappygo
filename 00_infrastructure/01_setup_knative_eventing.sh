#!/bin/bash

COL=70
SET_COL="echo -en \\033[${COL}G"
NORMAL="echo -en \\033[0;39m"
SUCCESS="echo -en \\033[1;32m"
FAILURE="echo -en \\033[1;31m"

print_status()
{
        if [ $# = 0 ]
        then
                echo "Usage: print_status {success|failure}"
                exit 1
        fi

        case "$1" in
                success)
                        $SET_COL
                        echo -n "[  "
                        $SUCCESS
                        echo -n "OK"
                        $NORMAL
                        echo "  ]"
                        ;;
                failure)
                        $SET_COL
                        echo -n "["
                        $FAILURE
                        echo -n "FAILED"
                        $NORMAL
                        echo "]"
                        ;;
        esac

}

RED='\033[0;31m'
GREEN='\033[0;32m'
LIGHTGREY='\033[0;37m'
NC='\033[0m' # No Color

# exit when any command fails
set -e

trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG

printf "${GREEN}"
echo "get credentials"
printf "${NC}"
printf "${LIGHTGREY}"
gcloud container clusters get-credentials knative-test
printf "${NC}"

printf "${GREEN}"
echo "Install the Eventing CRDs"
printf "${NC}"
printf "${LIGHTGREY}"
kubectl apply --selector knative.dev/crd-install=true \
--filename https://github.com/knative/eventing/releases/download/v0.11.0/release.yaml
printf "${NC}"

printf "${GREEN}"
echo "Install the Eventing sources"
printf "${NC}"
printf "${LIGHTGREY}"
kubectl apply --filename https://github.com/knative/eventing/releases/download/v0.11.0/release.yaml
printf "${NC}"

printf "${GREEN}"
echo "Confirm that Knative Eventing is correctly installed: "
printf "${NC}"
printf "${LIGHTGREY}"
kubectl get pods --namespace knative-eventing
printf "${NC}"

printf "${GREEN}"
echo "Create a namespace called event-example"
printf "${NC}"
printf "${LIGHTGREY}"
kubectl create namespace event-example
printf "${NC}"

printf "${GREEN}"
echo "gives the event-example namespace the knative-eventing-injection label, which adds resources that will allow you to manage your events."
printf "${NC}"
printf "${LIGHTGREY}"
kubectl label namespace event-example knative-eventing-injection=enabled
printf "${NC}"
