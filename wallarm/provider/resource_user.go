package wallarm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceWallarmUserCreate,
		Read:   resourceWallarmUserRead,
		Update: resourceWallarmUserUpdate,
		Delete: resourceWallarmUserDelete,

		Schema: map[string]*schema.Schema{
			"client_id": defaultClientIDWithValidationSchema,

			"email": {
				Type:     schema.TypeString,
				Required: true,
			},

			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: func(val interface{}, _ string) (warns []string, errs []error) {
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
				ValidateFunc: validation.StringInSlice([]string{"admin", "analytic", "auditor", "deploy", "partner_admin", "partner_admin_ext", "partner_analytic", "partner_auditor"}, false),
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
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
	clientID := retrieveClientID(d)
	email := d.Get("email").(string)
	realname := d.Get("realname").(string)
	permissions := d.Get("permissions").(string)
	enabled := d.Get("enabled").(bool)

	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	} else {
		password = passwordGenerate(10)
		d.Set("generated_password", password)
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

	d.Set("user_id", userID)

	resID := fmt.Sprintf("%d/%d", clientID, userID)
	d.SetId(resID)

	return resourceWallarmUserRead(d, m)
}
func resourceWallarmUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
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

	d.Set("realname", res.Body[0].Realname)
	d.Set("username", res.Body[0].Username)
	d.Set("enabled", res.Body[0].Enabled)
	d.Set("client_id", clientID)

	if len(res.Body[0].Permissions) > 0 {
		d.Set("permissions", res.Body[0].Permissions[0])
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
		userID := d.Get("user_id").(int)

		fields := &wallarm.UserFields{}
		hasChanges := false

		if d.HasChange("realname") {
			fields.Realname = d.Get("realname").(string)
			hasChanges = true
		}
		if d.HasChange("permissions") {
			fields.Permissions = []string{d.Get("permissions").(string)}
			hasChanges = true
		}

		if hasChanges {
			userBody := &wallarm.UserUpdate{
				UserFilter: &wallarm.UserFilter{
					ID: userID,
				},
				UserFields: fields,
			}
			if err := client.UserUpdate(userBody); err != nil {
				return err
			}
		}
	}

	return resourceWallarmUserRead(d, m)
}

func resourceWallarmUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(wallarm.API)
	clientID := retrieveClientID(d)
	userID := d.Get("user_id").(int)
	userBody := &wallarm.UserDelete{
		Filter: &wallarm.UserFilter{
			ID:       userID,
			Clientid: []int{clientID},
		},
	}
	if err := client.UserDelete(userBody); err != nil {
		return err
	}
	return nil
}
