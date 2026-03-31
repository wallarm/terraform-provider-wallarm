LOG_FILES = ./*.log
TMP_FILES = ./.DS_Store
TF_DIR = ./.terraform*
TF_STATE_FILE = ./terraform.tfstate*
IMPORT_BLOCKS_FILE = ./import_rule_blocks.tf
RULE_CONFIGS_FILE = ./import_rule_configs.tf
RULE_STAMP_CONFIGS_FILE = ./import_rule_stamp_configs.tf

FILES_DIRS_TO_CLEAN = ${LOG_FILES} ${TMP_FILES} ${TF_DIR} ${TF_STATE_FILE} ${IMPORT_BLOCKS_FILE} ${RULE_CONFIGS_FILE}


default: list

# ─── Basic operations ───────────────────────────────────────────────────────

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

destroy:
	@terraform destroy -refresh=false -auto-approve

clean:
	@if rm -rf ${FILES_DIRS_TO_CLEAN}; then echo "CLEAN OK"; else echo "CLEAN ERROR"; fi

# ─── Sync status ────────────────────────────────────────────────────────────
# Shows how many API rules are not in Terraform state. Read-only, no changes.

sync-status:
	@terraform plan -refresh=false -var 'sync_status=true'

# ─── Import building blocks ────────────────────────────────────────────────
# These targets are used individually or composed into the full import flows below.

# Generate import blocks via local_file resource (all rules)
import-blocks:
	@echo "Generating import blocks file..."
	@terraform apply -auto-approve -var='import_rules=true'

# Generate import blocks excluding stamp rules (for Case 2 split import)
import-blocks-no-stamps:
	@echo "Generating import blocks file (excluding stamps)..."
	@terraform apply -auto-approve -var='import_rules=true' -var='exclude_rule_types=["disable_stamp"]'

# Alternative: generate import blocks via terraform output instead of local_file
import-blocks-out:
	@echo "Generating import blocks file (via output)..."
	@terraform output -raw import_blocks > ${IMPORT_BLOCKS_FILE}

# Generate resource configs from import blocks using Terraform's native generator
import-configs:
	@echo "Generating resource configs from import blocks..."
	@terraform plan -generate-config-out=$(RULE_CONFIGS_FILE)

# Generate stamp-only import blocks + configs via wallarm_rule_generator
# Uses source="api" to handle complex point values (XML namespaces, etc.)
import-stamps-conf-gen:
	@echo "Generating stamp import blocks + configs via rule_generator..."
	@terraform apply -auto-approve -var='import_rules=true' -var='generate_configs=true' -var='rule_types=["disable_stamp"]'

# Fix generated configs: set defaults, remove null fields
# - variativity_disabled=false → true
# - comment=null → "Managed by Terraform"
# - remove all remaining =null lines
import-configs-fix:
	-@sed -E \
		-e 's/(variativity_disabled[[:space:]]*)=[[:space:]]*false/\1= true/' \
		-e 's/(comment[[:space:]]*)=[[:space:]]*null/\1= "Managed by Terraform"/' \
		-e '/=[[:space:]]*null/d' \
		$(RULE_CONFIGS_FILE) > $(RULE_CONFIGS_FILE).tmp && mv $(RULE_CONFIGS_FILE).tmp $(RULE_CONFIGS_FILE)

# ─── Case 1: Native import (all rules) ─────────────────────────────────────
# Uses -generate-config-out for all rule types.
# Works for most rules. May fail for disable_stamp rules with complex points.
#
# Steps: generate import blocks → generate configs → fix configs → apply

import-all-native: import-blocks import-configs
	@[ -f $(RULE_CONFIGS_FILE) ] && $(MAKE) import-configs-fix apply || \
		echo "\n $(RULE_CONFIGS_FILE) does not exist during execution, skipping import..."

# ─── Case 2: Split import (native + custom stamps) ─────────────────────────
# Step 1: Import all rules EXCEPT stamps via native -generate-config-out.
# Step 2: Import stamps via wallarm_rule_generator (handles complex points).
#
# Use when -generate-config-out fails on disable_stamp rules with long/special
# point values (e.g. XML namespace URIs with pipes, colons, semicolons).

# Step 1: native import for everything except stamps
import-all-no-stamps-native: import-blocks-no-stamps import-configs
	@[ -f $(RULE_CONFIGS_FILE) ] && $(MAKE) import-configs-fix apply || \
		echo "\n $(RULE_CONFIGS_FILE) does not exist during execution, skipping import..."

# Step 2: custom import for stamps only
import-stamps-custom: import-stamps-conf-gen import-blocks-out
	@[ -f $(RULE_STAMP_CONFIGS_FILE) ] && $(MAKE) import-configs-fix apply || \
		echo "\n $(RULE_STAMP_CONFIGS_FILE) does not exist during execution, skipping import..."
