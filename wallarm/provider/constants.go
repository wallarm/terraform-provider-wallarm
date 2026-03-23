package wallarm

// API pagination and batch size limits.
// All limits are centralized here to avoid scattering across files.
const (
	// IPListPageSize is the number of IP list groups fetched per API call.
	IPListPageSize = 1000

	// IPListMaxSubnets is the maximum number of subnet values allowed per IP list resource.
	IPListMaxSubnets = 1000

	// APIListLimit is the default limit for paginated API list requests (rules, users, etc.).
	APIListLimit = 500

	// HintBulkFetchLimit is the number of hints fetched per API call during cache bulk loading.
	HintBulkFetchLimit = 200

	// HintMaxBulkFetchPages caps the number of paginated requests to prevent runaway fetches.
	HintMaxBulkFetchPages = 500

	// HitFetchBatchSize is the number of hits fetched per API call.
	HitFetchBatchSize = 500

	// IPListCacheMaxRetries is the number of cache refresh retries after Create
	// to wait for API propagation.
	IPListCacheMaxRetries = 3

	// IPListCacheRetryDelay is the wait time between cache refresh retries.
	IPListCacheRetryDelay = 3
)
