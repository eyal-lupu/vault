package ldap

import (
	"strings"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathGroups(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `groups/(?P<name>.+)`,
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the LDAP group.",
			},

			"policies": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Comma-separated list of policies associated to the group.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathGroupDelete,
			logical.ReadOperation:   b.pathGroupRead,
			logical.UpdateOperation:  b.pathGroupWrite,
		},

		HelpSynopsis:    pathGroupHelpSyn,
		HelpDescription: pathGroupHelpDesc,
	}
}

func (b *backend) Group(s logical.Storage, n string) (*GroupEntry, error) {
	entry, err := s.Get("group/" + n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result GroupEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathGroupDelete(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete("group/" + d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathGroupRead(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	group, err := b.Group(req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"policies": strings.Join(group.Policies, ","),
		},
	}, nil
}

func (b *backend) pathGroupWrite(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	policies := strings.Split(d.Get("policies").(string), ",")
	for i, p := range policies {
		policies[i] = strings.TrimSpace(p)
	}

	// Store it
	entry, err := logical.StorageEntryJSON("group/"+name, &GroupEntry{
		Policies: policies,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	return nil, nil
}

type GroupEntry struct {
	Policies []string
}

const pathGroupHelpSyn = `
Manage users allowed to authenticate.
`

const pathGroupHelpDesc = `
This endpoint allows you to create, read, update, and delete configuration
for LDAP groups that are allowed to authenticate, and associate policies to
them.

Deleting a group will not revoke auth for prior authenticated users in that
group. To do this, do a revoke on "login/<username>" for
the usernames you want revoked.
`
