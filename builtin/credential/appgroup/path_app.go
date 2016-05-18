package appgroup

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/policyutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

// appStorageEntry stores all the options that are set on an App
type appStorageEntry struct {
	// Policies that are to be required by the token to access the App
	Policies []string `json:"policies" structs:"policies" mapstructure:"policies"`

	// Number of times the UserID generated against the App can be used to perform login
	NumUses int `json:"num_uses" structs:"num_uses" mapstructure:"num_uses"`

	// Duration (less than the backend mount's max TTL) after which a UserID generated against the App will expire
	UserIDTTL time.Duration `json:"userid_ttl" structs:"userid_ttl" mapstructure:"userid_ttl"`

	// Duration before which an issued token must be renewed
	TokenTTL time.Duration `json:"token_ttl" structs:"token_ttl" mapstructure:"token_ttl"`

	// Duration after which an issued token should not be allowed to be renewed
	TokenMaxTTL time.Duration `json:"token_max_ttl" structs:"token_max_ttl" mapstructure:"token_max_ttl"`

	// If set, activates cubbyhole mode for the UserIDs generated against the App.
	// An intermediary token will have the actual UserID reponse written in its cubbyhole.
	// The value of WrapTTL will be the duration after which the intermediary token
	// along with its cubbyhole will be destroyed.
	WrapTTL time.Duration `json:"wrap_ttl" structs:"wrap_ttl" mapstructure:"wrap_ttl"`
}

// appPaths creates all the paths that are used to register and manage an App.
//
// Paths returned:
// app/
// app/<app_name>
// app/policies
// app/num-uses
// app/userid-ttl
// app/token-ttl
// app/token-max-ttl
// app/wrap-ttl
// app/<app_name>/creds
// app/<app_name>/creds-specific
func appPaths(b *backend) []*framework.Path {
	return []*framework.Path{
		&framework.Path{
			Pattern: "app/?",
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathAppList,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-list"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-list"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name"),
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"policies": &framework.FieldSchema{
					Type:        framework.TypeString,
					Default:     "default",
					Description: "Comma separated list of policies on the App.",
				},
				"num_uses": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Number of times the a UserID can access the App, after which it will expire.",
				},
				"userid_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued UserID should expire.",
				},
				"token_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued token should expire.",
				},
				"token_max_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued token should not be allowed to be renewed.",
				},
				"wrap_ttl": &framework.FieldSchema{
					Type: framework.TypeDurationSecond,
					Description: `Duration in seconds, if specified, enables the Cubbyhole mode. In this mode,
the UserID creation endpoints will not return the UserID directly. Instead,
a new token will be returned with the UserID stored in its Cubbyhole. The
value of 'wrap_ttl' is the duration after which the returned token expires.
`,
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathAppCreateUpdate,
				logical.UpdateOperation: b.pathAppCreateUpdate,
				logical.ReadOperation:   b.pathAppRead,
				logical.DeleteOperation: b.pathAppDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/policies$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"policies": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Comma separated list of policies on the App.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppPoliciesUpdate,
				logical.ReadOperation:   b.pathAppPoliciesRead,
				logical.DeleteOperation: b.pathAppPoliciesDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-policies"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-policies"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/num-uses$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"num_uses": &framework.FieldSchema{
					Type:        framework.TypeInt,
					Description: "Number of times the a UserID can access the App, after which it will expire.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppNumUsesUpdate,
				logical.ReadOperation:   b.pathAppNumUsesRead,
				logical.DeleteOperation: b.pathAppNumUsesDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-num-uses"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-num-uses"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/userid-ttl$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"userid_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued UserID should expire.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppUserIDTTLUpdate,
				logical.ReadOperation:   b.pathAppUserIDTTLRead,
				logical.DeleteOperation: b.pathAppUserIDTTLDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-userid-ttl"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-userid-ttl"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/token-ttl$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"token_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued token should expire.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppTokenTTLUpdate,
				logical.ReadOperation:   b.pathAppTokenTTLRead,
				logical.DeleteOperation: b.pathAppTokenTTLDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-token-ttl"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-token-ttl"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/token-max-ttl$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"token_max_ttl": &framework.FieldSchema{
					Type:        framework.TypeDurationSecond,
					Description: "Duration in seconds after which the issued token should not be allowed to be renewed.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppTokenMaxTTLUpdate,
				logical.ReadOperation:   b.pathAppTokenMaxTTLRead,
				logical.DeleteOperation: b.pathAppTokenMaxTTLDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-token-max-ttl"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-token-max-ttl"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/wrap-ttl$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"wrap_ttl": &framework.FieldSchema{
					Type: framework.TypeDurationSecond,
					Description: `Duration in seconds, if specified, enables the Cubbyhole mode. In this mode,
the UserID creation endpoints will not return the UserID directly. Instead,
a new token will be returned with the UserID stored in its Cubbyhole. The
value of 'wrap_ttl' is the duration after which the returned token expires.
`,
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppWrapTTLUpdate,
				logical.ReadOperation:   b.pathAppWrapTTLRead,
				logical.DeleteOperation: b.pathAppWrapTTLDelete,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-wrap-ttl"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-wrap-ttl"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/creds$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"user_id": &framework.FieldSchema{
					Type:        framework.TypeString,
					Default:     "",
					Description: "NOT USER SUPPLIED AND UNDOCUMENTED.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation: b.pathAppCredsRead,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-creds"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-creds"][1]),
		},
		&framework.Path{
			Pattern: "app/" + framework.GenericNameRegex("app_name") + "/creds-specific$",
			Fields: map[string]*framework.FieldSchema{
				"app_name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the App.",
				},
				"user_id": &framework.FieldSchema{
					Type:        framework.TypeString,
					Default:     "",
					Description: "UserID to be attached to the App.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation: b.pathAppCredsSpecificUpdate,
			},
			HelpSynopsis:    strings.TrimSpace(appHelp["app-creds-specified"][0]),
			HelpDescription: strings.TrimSpace(appHelp["app-creds-specified"][1]),
		},
	}
}

// pathAppList is used to list all the Apps registered with the backend.
func (b *backend) pathAppList(
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.appLock.RLock()
	defer b.appLock.RUnlock()

	apps, err := req.Storage.List("app/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(apps), nil
}

// setAppEntry grabs a write lock and stores the options on an App into the storage
func (b *backend) setAppEntry(s logical.Storage, appName string, app *appStorageEntry) error {
	b.appLock.Lock()
	defer b.appLock.Unlock()
	if entry, err := logical.StorageEntryJSON("app/"+strings.ToLower(appName), app); err != nil {
		return err
	} else {
		return s.Put(entry)
	}
}

// appEntry grabs the read lock and fetches the options of an App from the storage
func (b *backend) appEntry(s logical.Storage, appName string) (*appStorageEntry, error) {
	if appName == "" {
		return nil, fmt.Errorf("missing app_name")
	}

	var result appStorageEntry

	b.appLock.RLock()
	defer b.appLock.RUnlock()

	if entry, err := s.Get("app/" + strings.ToLower(appName)); err != nil {
		return nil, err
	} else if entry == nil {
		return nil, nil
	} else if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// pathAppCreateUpdate registers a new App with the backend or updates the options
// of an existing App
func (b *backend) pathAppCreateUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	// Fetch or create an entry for the app
	app, err := b.appEntry(req.Storage, appName)
	if err != nil {
		return nil, err
	}
	// Create a new entry object if this is a CreateOperation
	if app == nil {
		app = &appStorageEntry{}
	}

	if policiesRaw, ok := data.GetOk("policies"); ok {
		app.Policies = policyutil.ParsePolicies(policiesRaw.(string))
	} else if req.Operation == logical.CreateOperation {
		app.Policies = policyutil.ParsePolicies(data.Get("policies").(string))
	}

	if numUsesRaw, ok := data.GetOk("num_uses"); ok {
		app.NumUses = numUsesRaw.(int)
	} else if req.Operation == logical.CreateOperation {
		app.NumUses = data.Get("num_uses").(int)
	}
	if app.NumUses < 0 {
		return logical.ErrorResponse("num_uses cannot be negative"), nil
	}

	if userIDTTLRaw, ok := data.GetOk("userid_ttl"); ok {
		app.UserIDTTL = time.Second * time.Duration(userIDTTLRaw.(int))
	} else if req.Operation == logical.CreateOperation {
		app.UserIDTTL = time.Second * time.Duration(data.Get("userid_ttl").(int))
	}

	if tokenTTLRaw, ok := data.GetOk("token_ttl"); ok {
		app.TokenTTL = time.Second * time.Duration(tokenTTLRaw.(int))
	} else if req.Operation == logical.CreateOperation {
		app.TokenTTL = time.Second * time.Duration(data.Get("token_ttl").(int))
	}

	if tokenMaxTTLRaw, ok := data.GetOk("token_max_ttl"); ok {
		app.TokenMaxTTL = time.Second * time.Duration(tokenMaxTTLRaw.(int))
	} else if req.Operation == logical.CreateOperation {
		app.TokenMaxTTL = time.Second * time.Duration(data.Get("token_max_ttl").(int))
	}

	// Check that the TokenMaxTTL value provided is less than the TokenMaxTTL.
	// Sanitizing the TTL and MaxTTL is not required now and can be performed
	// at credential issue time.
	if app.TokenMaxTTL > time.Duration(0) && app.TokenTTL > app.TokenMaxTTL {
		return logical.ErrorResponse("token_ttl should not be greater than token_max_ttl"), nil
	}

	// Update only if value is supplied. Defaults to zero.
	if wrapTTLRaw, ok := data.GetOk("wrap_ttl"); ok {
		app.WrapTTL = time.Duration(wrapTTLRaw.(int)) * time.Second
	}

	// Store the entry.
	return nil, b.setAppEntry(req.Storage, appName, app)
}

// pathAppRead grabs a read lock and reads the options set on the App from the storage
func (b *backend) pathAppRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		// Convert the 'time.Duration' values to second.
		app.UserIDTTL = app.UserIDTTL / time.Second
		app.TokenTTL = app.TokenTTL / time.Second
		app.TokenMaxTTL = app.TokenMaxTTL / time.Second
		app.WrapTTL = app.WrapTTL / time.Second

		return &logical.Response{
			Data: structs.New(app).Map(),
		}, nil
	}
}

// pathAppDelete removes the App from the storage
func (b *backend) pathAppDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}
	b.appLock.Lock()
	defer b.appLock.Unlock()

	return nil, req.Storage.Delete("app/" + strings.ToLower(appName))
}

func (b *backend) pathAppPoliciesUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if policiesRaw, ok := data.GetOk("policies"); ok {
		app.Policies = policyutil.ParsePolicies(policiesRaw.(string))
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing policies"), nil
	}
}

func (b *backend) pathAppPoliciesRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		return &logical.Response{
			Data: map[string]interface{}{
				"policies": app.Policies,
			},
		}, nil
	}
}

func (b *backend) pathAppPoliciesDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	// Deleting a field means resetting the value in the entry.
	app.Policies = (&appStorageEntry{}).Policies

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppNumUsesUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if numUsesRaw, ok := data.GetOk("num_uses"); ok {
		app.NumUses = numUsesRaw.(int)
		if app.NumUses < 0 {
			return logical.ErrorResponse("num_uses cannot be negative"), nil
		}
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing num_uses"), nil
	}
}

func (b *backend) pathAppNumUsesRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		return &logical.Response{
			Data: map[string]interface{}{
				"num_uses": app.NumUses,
			},
		}, nil
	}
}

func (b *backend) pathAppNumUsesDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	// Deleting a field means resetting the value in the entry.
	app.NumUses = (&appStorageEntry{}).NumUses

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppUserIDTTLUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if userIDTTLRaw, ok := data.GetOk("userid_ttl"); ok {
		app.UserIDTTL = time.Second * time.Duration(userIDTTLRaw.(int))
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing userid_ttl"), nil
	}
}

func (b *backend) pathAppUserIDTTLRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		app.UserIDTTL = app.UserIDTTL / time.Second
		return &logical.Response{
			Data: map[string]interface{}{
				"userid_ttl": app.UserIDTTL,
			},
		}, nil
	}
}

func (b *backend) pathAppUserIDTTLDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	// Deleting a field means resetting the value in the entry.
	app.UserIDTTL = (&appStorageEntry{}).UserIDTTL

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppTokenTTLUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if tokenTTLRaw, ok := data.GetOk("token_ttl"); ok {
		app.TokenTTL = time.Second * time.Duration(tokenTTLRaw.(int))
		if app.TokenMaxTTL > time.Duration(0) && app.TokenTTL > app.TokenMaxTTL {
			return logical.ErrorResponse("token_ttl should not be greater than token_max_ttl"), nil
		}
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing token_ttl"), nil
	}
}

func (b *backend) pathAppTokenTTLRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		app.TokenTTL = app.TokenTTL / time.Second
		return &logical.Response{
			Data: map[string]interface{}{
				"token_ttl": app.TokenTTL,
			},
		}, nil
	}
}

func (b *backend) pathAppTokenTTLDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	// Deleting a field means resetting the value in the entry.
	app.TokenTTL = (&appStorageEntry{}).TokenTTL

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppTokenMaxTTLUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if tokenMaxTTLRaw, ok := data.GetOk("token_max_ttl"); ok {
		app.TokenMaxTTL = time.Second * time.Duration(tokenMaxTTLRaw.(int))
		if app.TokenMaxTTL > time.Duration(0) && app.TokenTTL > app.TokenMaxTTL {
			return logical.ErrorResponse("token_max_ttl should be greater than or equal to token_ttl"), nil
		}
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing token_max_ttl"), nil
	}
}

func (b *backend) pathAppTokenMaxTTLRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		app.TokenMaxTTL = app.TokenMaxTTL / time.Second
		return &logical.Response{
			Data: map[string]interface{}{
				"token_max_ttl": app.TokenMaxTTL,
			},
		}, nil
	}
}

func (b *backend) pathAppTokenMaxTTLDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	// Deleting a field means resetting the value in the entry.
	app.TokenMaxTTL = (&appStorageEntry{}).TokenMaxTTL

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppWrapTTLUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	if wrapTTLRaw, ok := data.GetOk("wrap_ttl"); ok {
		app.WrapTTL = time.Duration(wrapTTLRaw.(int)) * time.Second
		return nil, b.setAppEntry(req.Storage, appName, app)
	} else {
		return logical.ErrorResponse("missing wrap_ttl"), nil
	}
}

func (b *backend) pathAppWrapTTLRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	if app, err := b.appEntry(req.Storage, strings.ToLower(appName)); err != nil {
		return nil, err
	} else if app == nil {
		return nil, nil
	} else {
		app.WrapTTL = app.WrapTTL / time.Second
		return &logical.Response{
			Data: map[string]interface{}{
				"wrap_ttl": app.WrapTTL,
			},
		}, nil
	}
}

func (b *backend) pathAppWrapTTLDelete(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}

	app.WrapTTL = (&appStorageEntry{}).WrapTTL

	return nil, b.setAppEntry(req.Storage, appName, app)
}

func (b *backend) pathAppCredsRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	userID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UserID:%s", err)
	}
	data.Raw["user_id"] = userID
	return b.handleAppCredsCommon(req, data, false)
}

func (b *backend) pathAppCredsSpecificUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	return b.handleAppCredsCommon(req, data, true)
}

func (b *backend) handleAppCredsCommon(req *logical.Request, data *framework.FieldData, specified bool) (*logical.Response, error) {
	appName := data.Get("app_name").(string)
	if appName == "" {
		return logical.ErrorResponse("missing app_name"), nil
	}

	userID := data.Get("user_id").(string)
	if userID == "" {
		return logical.ErrorResponse("missing user_id"), nil
	}

	app, err := b.appEntry(req.Storage, strings.ToLower(appName))
	if err != nil {
		return nil, err
	}
	if app == nil {
		return logical.ErrorResponse(fmt.Sprintf("app %s does not exist", appName)), nil
	}

	if err = b.registerUserIDEntry(req.Storage, selectorTypeApp, appName, userID, &userIDStorageEntry{
		NumUses:   app.NumUses,
		UserIDTTL: app.UserIDTTL,
	}); err != nil {
		return nil, fmt.Errorf("failed to store user ID: %s", err)
	}

	if specified {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"user_id": userID,
		},
	}, nil
}

var appHelp = map[string][2]string{
	"app-list": {
		"Lists all the Apps registered with the backend.",
		"The list will contain the names of the Apps.",
	},
	"app": {
		"Register an App with the backend.",
		`An App can represent a service, a machine or anything that can be IDed.
The set of policies on the App defines access to the App, meaning, any
Vault token with a policy set that is a superset of the policies on the
App registered here will have access to the App. If a UserID is desired
to be generated against only this specific App, it can be done via
'app/<app_name>/creds' and 'app/<app_name>/creds-specific' endpoints.
The properties of the UserID created against the App and the properties
of the token issued with the UserID generated againt the App, can be
configured using the parameters of this endpoint.`,
	},
	"app-policies": {
		"Policies of the App.",
		`A comma-delimited set of Vault policies that defines access to the App.
All the Vault tokens with policies that encompass the policy set
defined on the App, can access the App.`,
	},
	"app-num-uses": {
		"Use limit of the UserID generated against the App.",
		`If the UserIDs are generated/assigned against the App using the
'app/<app_name>/creds' or 'app/<app_name>/creds-specific' endpoints,
then the number of times that UserID can access the App is defined by
this option.`,
	},
	"app-userid-ttl": {
		`Duration in seconds, representing the lifetime of the UserIDs
that are generated against the App using 'app/<app_name>/creds' or
'app/<app_name>/creds-specific' endpoints.`,
		``,
	},
	"app-token-ttl": {
		`Duration in seconds, the lifetime of the token issued by using the UserID that
is generated against this App, before which the token needs to be renewed.`,
		`If UserIDs are generated against the App, using 'app/<app_name>/creds' or the
'app/<app_name>/creds-specific' endpoints, and if those UserIDs are used
to perform the login operation, then the value of 'token-ttl' defines the
lifetime of the token issued, before which the token needs to be renewed.`,
	},
	"app-token-max-ttl": {
		`Duration in seconds, the maximum lifetime of the tokens issued by using
the UserIDs that were generated against this App, after which the
tokens are not allowed to be renewed.`,
		`If UserIDs are generated against the App using 'app/<app_name>/creds'
or the 'app/<app_name>/creds-specific' endpoints, and if those UserIDs
are used to perform the login operation, then the value of 'token-max-ttl'
defines the maximum lifetime of the tokens issued, after which the tokens
cannot be renewed. A reauthentication is required after this duration.
This value will be capped by the backend mount's maximum TTL value.`,
	},
	"app-wrap-ttl": {
		"Duration in seconds, the lifetime of the wrapped token.",
		`Duration in seconds, if set, activates cubbyhole mode for the response.
In the cubbyhole mode, the generated UserID will not be returned as-is.
Instead, the response containing the UserID will be written in the
cubbyhole of a new token and this new token will be returned as a
response. The value of 'wrap_ttl' defines the lifetime of token which
contains the response in its cubbyhole.`,
	},
	"app-creds": {
		"Generate a UserID against this App.",
		`The UserID generated using this endpoint will be scoped to access
just this App and none else. The properties of this UserID will be
based on the options set on the App. It will expire after a period
defined by the 'userid_ttl' option on the App and/or the backend
mount's maximum TTL value.`,
	},
	"app-creds-specific": {
		"Assign a UserID of choice against the App.",
		`This option is not recommended unless there is a specific need
to do so. This will assign a client supplied UserID to be used to access
the App. This UserID will behave similarly to the UserIDs generated by
the backend. The properties of this UserID will be based on the options
set on the App. It will expire after a period defined by the 'userid_ttl'
option on the App and/or the backend mount's maximum TTL value.`,
	},
}