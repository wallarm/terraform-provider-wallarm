LOG_FILES = ./*.log
TMP_FILES = ./.DS_Store
TF_DIR = ./.terraform*
TF_STATE_FILE = ./terraform.tfstate*
RULES_DIR = ./configs/**
RULES_IMPORT_FILE = ./wallarm_rule_imports.tf
APP_IMPORT_FILE = ./wallarm_application_imports.tf
RESOURCES_FILE = ./imported_rules.tf

FILES_DIRS_TO_CLEAN = ${LOG_FILES} ${TMP_FILES} ${TF_DIR} ${TF_STATE_FILE} ${RULES_DIR} ${APP_IMPORT_FILE} ${RULES_IMPORT_FILE} ${RESOURCES_FILE}


default: state

state:
	@terraform state list

init:
	@echo "Initializing the Wallarm provider..."
	@terraform init

apply:
	@echo "Applying with automatic approval..."
	@terraform apply -auto-approve

clean:
	@if rm -rf ${FILES_DIRS_TO_CLEAN}; then echo "CLEAN OK"; else echo "CLEAN ERROR"; fi
