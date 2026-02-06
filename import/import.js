require('util').inspect.defaultOptions.depth = null; // DEBUGGING

const fs = require('node:fs');
const { createHash } = require('node:crypto');

const rulesJsonFile = process.argv[2];
const rulesImportFile = process.argv[3];

let data;

try {
  data = fs.readFileSync(rulesJsonFile, 'utf8');
} catch (err) {
  console.error(err.toString());
}

const wallarmRules = JSON.parse(data).body;

const WLRM_TF_RULE_TYPE = {
  disable_attack_type: 'wallarm_rule_disable_attack_type',
  parser_state: 'wallarm_rule_parser_state',
  disable_regex: 'wallarm_rule_ignore_regex',
  regex: 'wallarm_rule_regex',
  experimental_regex: 'wallarm_rule_regex', // Experimental
  binary_data: 'wallarm_rule_binary_data',
  uploads: 'wallarm_rule_uploads',
  overlimit_res_settings: 'wallarm_rule_overlimit_res_settings',
  vpatch: 'wallarm_rule_vpatch',
  attack_rechecker: 'wallarm_rule_attack_rechecker',
  attack_rechecker_rewrite: 'wallarm_rule_attack_rechecker_rewrite',
  set_response_header: 'wallarm_rule_set_response_header',
  wallarm_mode: 'wallarm_rule_mode',
  sensitive_data: 'wallarm_rule_masking',
  rate_limit: 'wallarm_rule_rate_limit',
  enum: 'wallarm_rule_enum',
  brute: 'wallarm_rule_brute',
  bola: 'wallarm_rule_bola',
  forced_browsing: 'wallarm_rule_forced_browsing',
  rate_limit_enum: 'wallarm_rule_rate_limit_enum',
  graphql_detection: 'wallarm_rule_graphql_detection',
  file_upload_size_limit: 'wallarm_rule_file_upload_size_limit',
  api_abuse_mode: null // Absent
}

const md5 = (str) => createHash('md5').update(str).digest('hex');

if (!wallarmRules) {
  console.log('No rules JSON data');
  return;
}

const wallarmRulesFiltered = wallarmRules.filter((r) => WLRM_TF_RULE_TYPE[r.type]);

const getIDTypePostfix = (rule) => {
  if (rule.wallarm_type === 'wallarm_mode') return `/${rule.mode}`;
  if (rule.wallarm_type === 'regex') return `/regex`;
  if (rule.wallarm_type === 'experimental_regex') return `/experimental_regex`;
  return '';
}

const terraformRules = wallarmRulesFiltered
  .map((r) => {
    r.wallarm_type = r.type;
    r.type = WLRM_TF_RULE_TYPE[r.type];
    r.terraform_id = `${r.clientid}/${r.actionid}/${r.id}${getIDTypePostfix(r)}`;
    r.terraform_name = `${r.wallarm_type}_${md5(String(r.id))}`;
    
    return r;
  });

const imports = terraformRules.map((r) => {
  return `import {
  to = ${r.type}.${r.terraform_name}
  id = "${r.terraform_id}"
}`
}).join('\n');

try {
  fs.writeFileSync(rulesImportFile, imports);
  console.log(`File ${rulesImportFile} was created successfully!`);
} catch (err) {
  console.error(err.toString());
}
