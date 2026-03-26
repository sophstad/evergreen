package graphql

import (
	"context"
	"fmt"

	"github.com/evergreen-ci/evergreen/model"
	model1 "github.com/evergreen-ci/evergreen/rest/model"
)

// Order is the resolver for the order field.
func (r *versionLiteResolver) Order(ctx context.Context, obj *model.Version) (int, error) {
	panic(fmt.Errorf("not implemented: Order - order"))
}

// Patch is the resolver for the patch field.
func (r *versionLiteResolver) Patch(ctx context.Context, obj *model.Version) (*model1.APIPatch, error) {
	panic(fmt.Errorf("not implemented: Patch - patch"))
}

// Project is the resolver for the project field.
func (r *versionLiteResolver) Project(ctx context.Context, obj *model.Version) (string, error) {
	panic(fmt.Errorf("not implemented: Project - project"))
}

// ProjectIdentifier is the resolver for the projectIdentifier field.
func (r *versionLiteResolver) ProjectIdentifier(ctx context.Context, obj *model.Version) (string, error) {
	panic(fmt.Errorf("not implemented: ProjectIdentifier - projectIdentifier"))
}

// UpstreamProject is the resolver for the upstreamProject field.
func (r *versionLiteResolver) UpstreamProject(ctx context.Context, obj *model.Version) (*UpstreamProject, error) {
	panic(fmt.Errorf("not implemented: UpstreamProject - upstreamProject"))
}

// User is the resolver for the user field.
func (r *versionLiteResolver) User(ctx context.Context, obj *model.Version) (*model1.APIDBUser, error) {
	panic(fmt.Errorf("not implemented: User - user"))
}

// VersionLite returns VersionLiteResolver implementation.
func (r *Resolver) VersionLite() VersionLiteResolver { return &versionLiteResolver{r} }

type versionLiteResolver struct{ *Resolver }
