TOKEN =
CLIENT_ID =
API_HOST = api.wallarm.com

RULES_FILE = rules_${CLIENT_ID}.json
IMPORT_FILE = rules_import.tf
RESOURCES_FILE = rules_resources_config.tf

OS := $(shell uname)

ifeq ($(OS), Darwin)
	SED_CMD = "sed -E -i '' '/((=[[:space:]]*null)|(client_id|mitigation|variativity_disabled)[[:space:]]*=)/d' $(RESOURCES_FILE)"
else
	SED_CMD = "sed -E -i '/((=[[:space:]]*null)|(client_id|mitigation|variativity_disabled)[[:space:]]*=)/d' $(RESOURCES_FILE)"
endif

default: info

info:
	@terraform state list
	@terraform plan

init:
	@echo "Initializing the Wallarm provider..."
	@terraform init

clean:
	@rm -rf *.log .DS_Store .terraform* terraform.tfstate* $(RULES_FILE) $(IMPORT_FILE) ${RESOURCES_FILE}
	@echo "CLEAN OK"

get-rules:
	@curl -v -X POST "https://$(API_HOST)/v1/objects/hint" -o $(RULES_FILE) \
		-H "X-WallarmApi-Token: $(TOKEN)" \
		-H "Content-Type: application/json" \
		-d "{ \"filter\": { \"clientid\": [ $(CLIENT_ID) ], \"system\": false }, \"offset\": 0, \"limit\": 1000}"
																### Use offset if there are more than 1000 rules

import: get-rules
	@node import.js $(RULES_FILE) $(IMPORT_FILE)
	@terraform plan -generate-config-out=$(RESOURCES_FILE)
	@eval $(SED_CMD)