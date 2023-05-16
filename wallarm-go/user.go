package wallarm

import (
	"encoding/json"
)

type (
	// User contains operations available on User resource
	User interface {
		UserRead(userBody *UserGet) (*UserRead, error)
		UserCreate(userBody *UserCreate) (*UserDetails, error)
		UserDelete(userBody *UserDelete) error
		UserUpdate(userBody *UserUpdate) error
		UserDetails() (*UserDetails, error)
	}

	// UserCreate is a POST body for the request to create a new user with the following parameters
	UserCreate struct {
		Email       string   `json:"email"`
		Phone       string   `json:"phone,omitempty"`
		Password    string   `json:"password"`
		Username    string   `json:"username"`
		Realname    string   `json:"realname"`
		Permissions []string `json:"permissions"`
		Clientid    int      `json:"clientid,omitempty"`
	}

	// UserRead is used as a response for the Read function for the User endpoints
	UserRead struct {
		Status int          `json:"status"`
		Body   []UserParams `json:"body"`
	}

	// UserParams wraps the entire parameters of the User entity
	UserParams struct {
		*UserCreate
		Phone             interface{} `json:"phone"`
		ID                int         `json:"id"`
		UUID              string      `json:"uuid"`
		ActualPermissions []string    `json:"actual_permissions"`
		MfaEnabled        bool        `json:"mfa_enabled"`
		CreateBy          int64       `json:"create_by"`
		CreateAt          int         `json:"create_at"`
		CreateFrom        string      `json:"create_from"`
		Enabled           bool        `json:"enabled"`
		Validated         bool        `json:"validated"`
		PasswordChanged   int         `json:"password_changed"`
		LoginHistory      []struct {
			Time int    `json:"time"`
			IP   string `json:"ip"`
		} `json:"login_history"`
		Timezone        string      `json:"timezone"`
		ResultsPerPage  int         `json:"results_per_page"`
		DefaultPool     string      `json:"default_pool"`
		DefaultPoolid   interface{} `json:"default_poolid"`
		LastReadNewsID  int         `json:"last_read_news_id"`
		SearchTemplates struct {
		} `json:"search_templates"`
		Notifications struct {
			ReportDaily struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"report_daily"`
			ReportWeekly struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"report_weekly"`
			ReportMonthly struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"report_monthly"`
			Vuln struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"vuln"`
			Scope struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"scope"`
			System struct {
				Email bool `json:"email"`
				Sms   bool `json:"sms"`
			} `json:"system"`
		} `json:"notifications"`
		Components               []string    `json:"components"`
		Language                 string      `json:"language"`
		LastLoginTime            int         `json:"last_login_time"`
		DateFormat               string      `json:"date_format"`
		TimeFormat               string      `json:"time_format"`
		JobTitle                 interface{} `json:"job_title"`
		AvailableAuthentications []string    `json:"available_authentications"`
		FrontendURL              string      `json:"frontend_url"`
	}

	// UserGet is used for Read function purposes (to request Users by a GET method)
	UserGet struct {
		Limit     int         `json:"limit"`
		OrderBy   string      `json:"order_by"`
		OrderDesc bool        `json:"order_desc"`
		Filter    *UserFilter `json:"filter"`
	}

	// UserDetails is used as a response for request about the specific User.
	// For example, it may be used to find out a parameter (Client ID) for the current user which auth params are used
	UserDetails struct {
		Status int         `json:"status"`
		Body   *UserParams `json:"body"`
	}

	// UserFilter is intended to filter Users for the Delete purpose
	UserFilter struct {
		ID       int    `json:"id,omitempty"`
		Clientid int    `json:"clientid,omitempty"`
		UUID     string `json:"uuid,omitempty"`
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
	}

	// UserDelete is utilised to Delete Users
	UserDelete struct {
		Filter *UserFilter `json:"filter"`
	}

	// UserFields represents fields of the Users for Update function
	UserFields struct {
		Phone          string   `json:"phone,omitempty"`
		Realname       string   `json:"realname,omitempty"`
		Permissions    []string `json:"permissions,omitempty"`
		Clientid       int      `json:"clientid,omitempty"`
		Timezone       string   `json:"timezone,omitempty"`
		JobTitle       string   `json:"job_title,omitempty"`
		ResultsPerPage int      `json:"results_per_page,omitempty"`
		DefaultPool    string   `json:"default_pool,omitempty"`
		Enabled        bool     `json:"enabled,omitempty"`
		Password       string   `json:"password,omitempty"`
		Notifications  struct {
			ReportDaily struct {
				Email bool `json:"email,omitempty"`
				Sms   bool `json:"sms,omitempty"`
			} `json:"report_daily,omitempty"`
			ReportWeekly struct {
				Email bool `json:"email,omitempty"`
				Sms   bool `json:"sms,omitempty"`
			} `json:"report_weekly,omitempty"`
			ReportMonthly struct {
				Email bool `json:"email,omitempty"`
				Sms   bool `json:"sms,omitempty"`
			} `json:"report_monthly,omitempty"`
			System struct {
				Email bool `json:"email,omitempty"`
				Sms   bool `json:"sms,omitempty"`
			} `json:"system,omitempty"`
			Vuln struct {
				Email bool `json:"email,omitempty"`
				Sms   bool `json:"sms,omitempty"`
			} `json:"vuln,omitempty"`
		} `json:"notifications,omitempty"`
		SearchTemplates struct {
		} `json:"search_templates,omitempty"`
	}

	// UserUpdate is used to Update Users
	UserUpdate struct {
		*UserFilter `json:"filter,omitempty"`
		*UserFields `json:"fields,omitempty"`
		Limit       int    `json:"limit,omitempty"`
		Offset      int    `json:"offset,omitempty"`
		OrderBy     string `json:"order_by,omitempty"`
		OrderDesc   bool   `json:"order_desc,omitempty"`
	}
)

// UserRead to read the particular user defined by a body
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) UserRead(userBody *UserGet) (*UserRead, error) {

	uri := "/v1/objects/user"
	respBody, err := api.makeRequest("POST", uri, "user", userBody)
	if err != nil {
		return nil, err
	}
	var u UserRead
	if err = json.Unmarshal(respBody, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// UserCreate demonstrates the function to create a new user
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) UserCreate(userBody *UserCreate) (*UserDetails, error) {

	uri := "/v1/objects/user/create"
	respBody, err := api.makeRequest("POST", uri, "user", userBody)
	if err != nil {
		return nil, err
	}
	var u UserDetails
	if err = json.Unmarshal(respBody, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// UserDelete is due to delete Users defined by filter, for example, `email`
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) UserDelete(userBody *UserDelete) error {

	uri := "/v1/objects/user/delete"
	_, err := api.makeRequest("POST", uri, "user", userBody)
	if err != nil {
		return err
	}
	return nil
}

// UserUpdate is due to update some User fields such as (phone, realname, password and notification state)
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) UserUpdate(userBody *UserUpdate) error {

	uri := "/v1/objects/user/update"
	_, err := api.makeRequest("POST", uri, "user", userBody)
	if err != nil {
		return err
	}
	return nil
}

// UserDetails is used to request characteristics of the user.
// API reference: https://apiconsole.eu1.wallarm.com
func (api *api) UserDetails() (*UserDetails, error) {

	uri := "/v1/user"
	respBody, err := api.makeRequest("POST", uri, "userdetails", nil)
	if err != nil {
		return nil, err
	}
	var u UserDetails
	if err = json.Unmarshal(respBody, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
