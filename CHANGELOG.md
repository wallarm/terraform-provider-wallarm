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
