#!/bin/bash

SCRIPTPATH="$( cd "$(dirname scripts/plugindircheck.sh)" >/dev/null 2>&1 ; pwd -P )"

### Check for dir, if not found create it using the mkdir ##
if [ -d "~/.terraform.d/plugins/$1_$2" ]; then
    cp $SCRIPTPATH/../terraform-provider-wallarm_$2 ~/.terraform.d/plugins/$1_$2/terraform-provider-wallarm_$3
else
    mkdir -p ~/.terraform.d/plugins/$1_$2
    cp $SCRIPTPATH/../terraform-provider-wallarm_$3 ~/.terraform.d/plugins/$1_$2/terraform-provider-wallarm_$3
fi