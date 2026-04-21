package wallarm

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/wallarm/wallarm-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceWallarmUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWallarmUserCreate,
		ReadContext:   resourceWallarmUserRead,
		UpdateContext: resourceWallarmUserUpdate,
		DeleteContext: resourceWallarmUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWallarmUserImport,
		},

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
				ValidateFunc: validation.StringInSlice([]string{"admin", "admin_ext", "analytic", "auditor", "deploy", "partner_admin", "partner_admin_ext", "partner_analytic", "partner_auditor"}, false),
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

func resourceWallarmUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	email := d.Get("email").(string)
	realname := d.Get("realname").(string)
	permissions := d.Get("permissions").(string)
	enabled := d.Get("enabled").(bool)

	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	} else {
		generated, err := passwordGenerate(10)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to generate password: %w", err))
		}
		password = generated
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
			return diag.FromErr(ImportAsExistsError("wallarm_user", existingID))
		}
		return diag.FromErr(err)
	}

	userID := res.Body.ID

	d.Set("user_id", userID)

	resID := fmt.Sprintf("%d/%d", clientID, userID)
	d.SetId(resID)

	return resourceWallarmUserRead(ctx, d, m)
}
func resourceWallarmUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	userID := d.Get("user_id").(int)
	user := &wallarm.UserGet{
		Limit:     APIListLimit,
		OrderBy:   "realname",
		OrderDesc: false,
		Filter: &wallarm.UserFilter{
			ID: userID,
		},
	}
	res, err := client.UserRead(user)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(res.Body) == 0 {
		body, err := json.Marshal(res)
		if err != nil {
			return diag.FromErr(err)
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

func resourceWallarmUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("email") || d.HasChange("password") {
		if diags := resourceWallarmUserDelete(ctx, d, m); diags != nil {
			return diags
		}
		if diags := resourceWallarmUserCreate(ctx, d, m); diags != nil {
			return diags
		}
	} else {
		client := apiClient(m)
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
				return diag.FromErr(err)
			}
		}
	}

	return resourceWallarmUserRead(ctx, d, m)
}

func passwordGenerate(length int) (string, error) {
	digits := "0123456789"
	specials := "~=+%^*()_[]{}!@#$?"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	var err error
	if buf[0], err = cryptoRandByte(digits); err != nil {
		return "", err
	}
	if buf[1], err = cryptoRandByte(specials); err != nil {
		return "", err
	}
	for i := 2; i < length; i++ {
		if buf[i], err = cryptoRandByte(all); err != nil {
			return "", err
		}
	}
	for i := len(buf) - 1; i > 0; i-- {
		j, err := cryptoRandIntn(i + 1)
		if err != nil {
			return "", err
		}
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf), nil
}

func cryptoRandByte(charset string) (byte, error) {
	idx, err := cryptoRandIntn(len(charset))
	if err != nil {
		return 0, err
	}
	return charset[idx], nil
}

func cryptoRandIntn(n int) (int, error) {
	maxN := big.NewInt(int64(n))
	v, err := crand.Int(crand.Reader, maxN)
	if err != nil {
		return 0, fmt.Errorf("crypto/rand failed: %w", err)
	}
	return int(v.Int64()), nil
}

func isPasswordValid(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func resourceWallarmUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := apiClient(m)
	clientID, err := retrieveClientID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	userID := d.Get("user_id").(int)
	userBody := &wallarm.UserDelete{
		Filter: &wallarm.UserFilter{
			ID:       userID,
			Clientid: []int{clientID},
		},
	}
	if err := client.UserDelete(userBody); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// resourceWallarmUserImport parses a 2-part import ID "{client_id}/{user_id}".
func resourceWallarmUserImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid id (%q) specified, should be in format \"{client_id}/{user_id}\"", d.Id())
	}
	clientID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid client_id: %w", err)
	}
	userID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	d.Set("client_id", clientID)
	d.Set("user_id", userID)
	d.SetId(fmt.Sprintf("%d/%d", clientID, userID))
	return []*schema.ResourceData{d}, nil
}
