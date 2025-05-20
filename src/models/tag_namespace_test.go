package models_test

import (
	"testing"
	"entitydb/models"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag    string
		want   *models.TagNamespace
	}{
		{
			tag: "rbac:perm:entity:create",
			want: &models.TagNamespace{
				Namespace: "rbac",
				Path:      []string{"perm", "entity"},
				Value:     "create",
			},
		},
		{
			tag: "type:user",
			want: &models.TagNamespace{
				Namespace: "type",
				Path:      []string{},
				Value:     "user",
			},
		},
		{
			tag: "id:username:admin",
			want: &models.TagNamespace{
				Namespace: "id",
				Path:      []string{"username"},
				Value:     "admin",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := models.ParseTag(tt.tag)
			if got == nil && tt.want != nil {
				t.Errorf("ParseTag() = nil, want %v", tt.want)
				return
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("ParseTag().Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if len(got.Path) != len(tt.want.Path) {
				t.Errorf("ParseTag().Path length = %v, want %v", len(got.Path), len(tt.want.Path))
			}
			if got.Value != tt.want.Value {
				t.Errorf("ParseTag().Value = %v, want %v", got.Value, tt.want.Value)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name         string
		tags         []string
		requiredPerm string
		want         bool
	}{
		{
			name:         "exact match",
			tags:         []string{"rbac:perm:entity:create", "rbac:perm:entity:read"},
			requiredPerm: "rbac:perm:entity:create",
			want:         true,
		},
		{
			name:         "wildcard all",
			tags:         []string{"rbac:perm:*"},
			requiredPerm: "rbac:perm:entity:create",
			want:         true,
		},
		{
			name:         "wildcard category",
			tags:         []string{"rbac:perm:entity:*"},
			requiredPerm: "rbac:perm:entity:create",
			want:         true,
		},
		{
			name:         "no match",
			tags:         []string{"rbac:perm:entity:read"},
			requiredPerm: "rbac:perm:entity:create",
			want:         false,
		},
		{
			name:         "wildcard doesn't match different category",
			tags:         []string{"rbac:perm:entity:*"},
			requiredPerm: "rbac:perm:issue:create",
			want:         false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := models.HasPermission(tt.tags, tt.requiredPerm); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTagsByNamespace(t *testing.T) {
	tags := []string{
		"type:user",
		"id:username:admin",
		"rbac:role:admin",
		"rbac:perm:*",
		"status:active",
	}
	
	rbacTags := models.GetTagsByNamespace(tags, "rbac")
	if len(rbacTags) != 2 {
		t.Errorf("Expected 2 rbac tags, got %d", len(rbacTags))
	}
	
	idTags := models.GetTagsByNamespace(tags, "id")
	if len(idTags) != 1 {
		t.Errorf("Expected 1 id tag, got %d", len(idTags))
	}
}