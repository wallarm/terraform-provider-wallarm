LOG_FILES = ./*.log
TMP_FILES = ./.DS_Store
TF_DIR = ./.terraform*
TF_STATE_FILE = ./terraform.tfstate*
IMPORT_BLOCKS_FILE = ./import_rule_blocks.tf
RESOURCES_FILE = ./import_rule_configs.tf

FILES_DIRS_TO_CLEAN = ${LOG_FILES} ${TMP_FILES} ${TF_DIR} ${TF_STATE_FILE} ${IMPORT_BLOCKS_FILE} ${RESOURCES_FILE}


default: list

list:
	@terraform state list

init:
	@echo "Initializing the Wallarm provider..."
	@terraform init

plan:
	@terraform plan

apply:
	@echo "Applying with automatic approval..."
	@terraform apply -auto-approve

sync-status:
	@terraform plan -var 'sync_status=true'

import-blocks:
	@echo "Generating import blocks file..."
	@terraform apply -auto-approve -var='import_rules=true'

import-configs:
	@echo "Generating configuration file for rules resources import..."
	@terraform plan -generate-config-out=$(RESOURCES_FILE)

import-configs-fix: 
	-@sed -E '/((=[[:space:]]*null)|(mitigation|variativity_disabled)[[:space:]]*=)/d' $(RESOURCES_FILE) > $(RESOURCES_FILE).tmp && \
		mv $(RESOURCES_FILE).tmp $(RESOURCES_FILE)

import: import-blocks import-configs 
	@[ -f $(RESOURCES_FILE) ] && $(MAKE) import-configs-fix apply || \
		echo "\n $(RESOURCES_FILE) does not exist during execution, skipping import..."

destroy:
	@terraform destroy -auto-approve

clean:
	@if rm -rf ${FILES_DIRS_TO_CLEAN}; then echo "CLEAN OK"; else echo "CLEAN ERROR"; fi
