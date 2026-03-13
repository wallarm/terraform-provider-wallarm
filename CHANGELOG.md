# v2.1.0 (Mar 17, 2026)

## NOTES:

* Completed migration to Terraform Plugin SDKv2
* Added IP Lists Auto Import
* Improved logging system
* Implemented APIError type
* Added errors handling for 5xx and some 4xx responses
* Added 'attack' mode to fetch hits by data source
* Fixed bugs
* Updated tests
* Updated documentation
* Added a module for interaction with Wallarm API: `./examples/terraform-wallarm-api`

# v2.0.0 (Mar 12, 2026)

## NOTES:

* Updated Terraform Plugin SDK to v2
* Updated IP List to support all current API source types
* Updated existing integrations
* Removed Scanner
* Added data_source_rules for bulk rules imports
* Added data_source_applications for bulk applications imports
* Added retries for rules CREATE/UPDATE/DELETE methods to handle snapshot 423 error
* Fixed tests
* Fixed bugs

# v1.9.0 (Mar 6, 2026)

## NOTES:

* Added new data source `data_source_wallarm_hits.go`
* Added middleware between API and terraform provider `hint_cache.go`
* Bug fixes

# v1.8.6 (Feb 12, 2026)

## NOTES:

* Added new resource `wallarm_rule_disable_stamp`

# v1.8.5 (Feb 06, 2026)

## NOTES:

* Add updating comment and variativity_disabled
* Patch wrapPointElements -> support gql params

# v1.8.4 (Jan 28, 2026)

## NOTES:

* Resource rule cred stuff is removed

# v1.8.3 (Jan 28, 2026)

## NOTES:

* Resource rules import refactored
* Go mod updated 1.23 -> 1.24

# v1.8.2 (Jul 24, 2025)

## NOTES:

* Added new rules
* Code refactored

# v1.8.1 (Jul 24, 2025)

## NOTES:

* Added fields 'set', 'active', 'title', 'mitigation' for all resource_rules
* Code refactored

# v1.8.0 (April 17, 2025)

## NOTES:

* Code refactored
* Go mod updated 1.15 -> 1.23

# v1.7.0 (April 15, 2025)

## NOTES:

* Updated event_type: 'vuln' -> 'vuln_high' in docs/resources
* Removed unused resources 'attack_rechecker', 'attack_rechecker_rewrite'
* Extended validation for resourceWallarmTrigger, added check 'forced_browsing_started'
* Fixed lock_time for triggers, now it fills only for action_id 'block_ips' or 'add_to_graylist'

# v1.6.0 (December 6, 2024)

## NOTES:

* Added import support
* Added OverlimitResSettingsRule resource
* Changed rule vpatch resource according to api
* Changed rule set response header resource according to api
* Fixed some api methods

# v1.5.0 (September 1, 2024)

## NOTES:

* Added integration with DataDog
* Added integration with Telegram
* Added integration with MS Teams
* Added API Spec resource for managing API specifications
* Added Rate Limit rule
* Added support for all triggers that we have in the UI

# v1.1.0 (May 24, 2023)

## BUG FIXES:

* Fixed intergation api methods
* Fixed rule api methods
* Fixed trigger api methods

## NOTES:

* Added support for the following new resources: `wallarm_allowlist`, `wallarm_graylist`
* Added support of the query field in rules
* Change api authentication method using X-WallarmAPI-Token

# v1.0.0 (November 3, 2022)

## BUG FIXES:

* Fixed the following resources that were broken by changes in the API: `wallarm_blacklist`, `wallarm_rule_bruteforce_counter`, `wallarm_rule_dirbust_counter`, `wallarm_global_mode`
* Fixed various other bugs

## NOTES:

* Added support for the following new resources: `wallarm_rule_binary_data`, `wallarm_rule_disable_attack_type`, `wallarm_rule_parser_state`, `wallarm_rule_uploads`, `wallarm_rule_variative_keys`, `wallarm_rule_variative_values`
* Renamed the resource `wallarm_blacklist` to `wallarm_denylist`
* Updated the docs

# v0.0.10 (September 10, 2021)

## BUG FIXES:

* Fixed Opsgenie integration resource

## NOTES:

* Minor changes in the dependencies
* Fixed broken links and typos in the docs

# v0.0.9 (January 8, 2021)

## NOTES:

* Fix bug with the incorrect state
* Now `PASSPHRASE` for Release workflow is taken from `secrets.PASSPHRASE` as per best practices
* Added support of `go 1.15` in tests
* Two newly created resources: `wallarm_rule_bruteforce_counter` and `wallarm_rule_dirbust_counter`. The source code and documentation have been included altogether.
* `Trigger` arbitrary time value replaced by enum type
* New `wallarm-go` library structure approach

# v0.0.8 (December 16, 2020)

## NOTES:

* Minor changes in markdown

# v0.0.7 (October 23, 2020)

## NOTES:

* Documentation layout have been modified

# v0.0.3 (October 10, 2020)

## NOTES:

* Automatic GO release have been added

# v0.0.2 (September 16, 2020)

## FEATURES:

**New Resource:** `wallarm_blacklist` - is used to manage the blacklist

## ENHANCEMENTS:

**Resource:** `wallarm_rule_*` - bumped a client library version to comply with the new struct for `[][]interface{}`

## BUG FIXES:

**Resource:** - fixed 400 HTTP response code when there is an incorrect JSON body for requests on the update call

# v0.0.1 (September 10, 2020)

## NOTES:

* The first public release
