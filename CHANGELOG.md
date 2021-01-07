# v0.0.11 (January 2, 2021)

## NOTES:

* Fix bug with the incorrect state
* Now `PASSPHRASE` for Release workflow is taken from `secrets.PASSPHRASE` as per best practices
* Added support of `go 1.15` in tests
* Two newly created resources: `wallarm_rule_bruteforce_counter` and `wallarm_rule_dirbust_counter`. The source code and documentation have been included altogether.
* `Trigger` arbitrary time value replaced by enum type
* New `wallarm-go` library structure approach

# v0.0.8 (December 2, 2020)

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