LOG_FILES = ./*.log
TMP_FILES = ./.DS_Store
TF_DIR = ./.terraform*
TF_STATE_FILE = ./terraform.tfstate*
RULE_CONFIGS_DIR = ./generated_rules

FILES_DIRS_TO_CLEAN = ${LOG_FILES} ${TMP_FILES} ${TF_DIR} ${TF_STATE_FILE} ${RULE_CONFIGS_DIR}


default: list

list:
	@terraform state list

init:
	@echo "Initializing the Wallarm provider..."
	@terraform init && terraform apply -auto-approve -var='request_ids={}'

plan:
	@terraform plan

apply:
	@echo "Applying with automatic approval..."
	@terraform apply -auto-approve

destroy:
	@terraform destroy -refresh=false -auto-approve

clean:
	@if rm -rf ${FILES_DIRS_TO_CLEAN}; then echo "CLEAN OK"; else echo "CLEAN ERROR"; fi
