LOG_FILES = ./*.log
TMP_FILES = ./.DS_Store
TF_DIR = ./.terraform*
TF_STATE_FILE = ./terraform.tfstate*
FP_RULES_DIR = ./fp-rules-configs
CUSTOM_RULES_DIR = ./custom-rules-configs
IMPORT_RULES_DIR = ./import-rules-configs
IMPORT_FILES = ./rules_import_blocks.tf
APP_IMPORTS = ./wallarm_application_imports.tf
RESOURCES_FILE = ./rules_import_resources_configs.tf

FILES_DIRS_TO_CLEAN = ${LOG_FILES} ${TMP_FILES} ${TF_DIR} ${TF_STATE_FILE} ${FP_RULES_DIR} ${CUSTOM_RULES_DIR} ${IMPORT_RULES_DIR} ${APP_IMPORTS} ${IMPORT_FILE} ${RESOURCES_FILE}


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
