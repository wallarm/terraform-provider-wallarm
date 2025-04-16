package wallarm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceWallarmUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmUserCreate,
		Read:   resourceWallarmUserRead,
		Update: resourceWallarmUserUpdate,
		Delete: resourceWallarmUserDelete,

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Client ID to perform changes",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v <= 0 {
						errs = append(errs, fmt.Errorf("%q must be positive, got: %d", key, v))
					}
					return
				},
			},

			"email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if isPasswordValid(v) {
						return
					}
					errs = append(errs, fmt.Errorf("use at least 8 characters containing at least 1 number, 1 special character, 1 lowercase letter and 1 uppercase letter"))
					return
				},
			},

			"permissions": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"admin", "analyst", "deploy", "read_only", "global_admin", "global_analyst", "global_read_only"}, false),
			},

			"realname": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\w+\s+\w+`), "There should be two words separated by one or more spaces"),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"generated_password": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceWallarmUserCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	email := d.Get("email").(string)
	realname := d.Get("realname").(string)
	permissions := d.Get("permissions").(string)
	enabled := d.Get("enabled").(bool)

	switch permissions {
	case "analyst":
		permissions = "analytic"
	case "read_only":
		permissions = "auditor"
	case "global_read_only":
		permissions = "partner_auditor"
	case "global_analyst":
		permissions = "partner_analytic"
	case "global_admin":
		permissions = "partner_admin"
	}

	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	} else {
		password = passwordGenerate(10)
		if err := d.Set("generated_password", password); err != nil {
			return err
		}
	}

	userBody := &wallarm.UserCreate{
		Email:       email,
		Password:    password,
		Username:    email,
		Realname:    realname,
		Permissions: []string{permissions},
		Clientid:    clientID,
		Enabled:     enabled,
	}

	res, err := client.UserCreate(userBody)
	if err != nil {
		if errors.Is(err, wallarm.ErrExistingResource) {
			existingID := fmt.Sprintf("%d/%s", clientID, email)
			return ImportAsExistsError("wallarm_user", existingID)
		}
		return err
	}

	userID := res.Body.ID

	if err := d.Set("user_id", userID); err != nil {
		return err
	}

	resID := fmt.Sprintf("%d/%s/%d", clientID, realname, userID)
	d.SetId(resID)

	return resourceWallarmUserRead(d, m)
}
func resourceWallarmUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d, client)
	userID := d.Get("user_id").(int)
	user := &wallarm.UserGet{
		Limit:     1000,
		OrderBy:   "realname",
		OrderDesc: false,
		Filter: &wallarm.UserFilter{
			ID: userID,
		},
	}
	res, err := client.UserRead(user)
	if err != nil {
		return err
	}

	if len(res.Body) == 0 {
		body, err := json.Marshal(res)
		if err != nil {
			return err
		}
		log.Printf("[WARN] User hasn't been found in API. Body: %s", body)

		d.SetId("")
		return nil
	}

	if err = d.Set("realname", res.Body[0].Realname); err != nil {
		return err
	}

	if err = d.Set("username", res.Body[0].Username); err != nil {
		return err
	}

	if err = d.Set("enabled", res.Body[0].Enabled); err != nil {
		return err
	}

	if err = d.Set("client_id", clientID); err != nil {
		return err
	}

	return nil
}

func resourceWallarmUserUpdate(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("email") || d.HasChange("password") {
		if err := resourceWallarmUserDelete(d, m); err != nil {
			return err
		}
		if err := resourceWallarmUserCreate(d, m); err != nil {
			return err
		}
	} else {
		client := m.(wallarm.API)
		clientID := retrieveClientID(d, client)
		email := d.Get("email").(string)
		realname := d.Get("realname").(string)
		permissions := d.Get("permissions").(string)
		enabled := d.Get("enabled").(bool)
		var password string
		if d.HasChange("password") {
			if v, ok := d.GetOk("password"); ok {
				password = v.(string)
			} else {
				password = passwordGenerate(10)
			}
		}
		userBody := &wallarm.UserUpdate{
			UserFilter: &wallarm.UserFilter{
				Email:    email,
				Username: email,
			},
			UserFields: &wallarm.UserFields{
				Password:    password,
				Realname:    realname,
				Permissions: []string{permissions},
				Enabled:     enabled,
				Clientid:    clientID,
			},
			Limit: 1000,
		}
		if err := client.UserUpdate(userBody); err != nil {
			return err
		}

	}

	return resourceWallarmUserRead(d, m)
}

func resourceWallarmUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	userID := d.Get("user_id").(int)
	userBody := &wallarm.UserDelete{
		Filter: &wallarm.UserFilter{
			ID: userID}}
	if err := client.UserDelete(userBody); err != nil {
		return err
	}
	return nil
}
