package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/user"
	restModel "github.com/evergreen-ci/evergreen/rest/model"
	"github.com/evergreen-ci/utility"
)

// GitTags is the resolver for the gitTags field.
func (r *activeVersionResolver) GitTags(ctx context.Context, obj *model.ActiveWaterfallVersion) ([]*restModel.APIGitTag, error) {
	apiTags := make([]*restModel.APIGitTag, 0, len(obj.GitTags))
	for _, gt := range obj.GitTags {
		apiTags = append(apiTags, &restModel.APIGitTag{
			Tag:    utility.ToStringPtr(gt.Tag),
			Pusher: utility.ToStringPtr(gt.Pusher),
		})
	}
	return apiTags, nil
}

// User is the resolver for the user field.
func (r *activeVersionResolver) User(ctx context.Context, obj *model.ActiveWaterfallVersion) (*restModel.APIDBUser, error) {
	requestedFields := graphql.CollectAllFields(ctx)
	needsDBFetch := false
	for _, field := range requestedFields {
		if field != "userId" && field != "displayName" && field != "emailAddress" {
			needsDBFetch = true
			break
		}
	}

	if !needsDBFetch {
		return &restModel.APIDBUser{
			UserID:       utility.ToStringPtr(obj.AuthorID),
			DisplayName:  utility.ToStringPtr(obj.Author),
			EmailAddress: utility.ToStringPtr(obj.AuthorEmail),
		}, nil
	}

	authorId := obj.AuthorID
	currentUser := mustHaveUser(ctx)
	if currentUser.Id == authorId {
		apiUser := &restModel.APIDBUser{}
		apiUser.BuildFromService(*currentUser)
		return apiUser, nil
	}

	author, err := user.FindOneById(ctx, authorId)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("getting user '%s': %s", authorId, err.Error()))
	}
	if author == nil {
		return &restModel.APIDBUser{
			UserID:       utility.ToStringPtr(obj.AuthorID),
			DisplayName:  utility.ToStringPtr(obj.Author),
			EmailAddress: utility.ToStringPtr(obj.AuthorEmail),
		}, nil
	}

	apiUser := &restModel.APIDBUser{}
	apiUser.BuildFromService(*author)
	return apiUser, nil
}

// WaterfallBuilds is the resolver for the waterfallBuilds field.
func (r *activeVersionResolver) WaterfallBuilds(ctx context.Context, obj *model.ActiveWaterfallVersion) ([]*model.WaterfallBuild, error) {
	builds, err := model.GetVersionBuilds(ctx, obj.Id, obj.BuildIds)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("getting build variants for version '%s': %s", obj.Id, err.Error()))
	}
	versionBuilds := make([]*model.WaterfallBuild, len(builds))
	for i := range builds {
		versionBuilds[i] = &builds[i]
	}
	return versionBuilds, nil
}

// ActiveVersion returns ActiveVersionResolver implementation.
func (r *Resolver) ActiveVersion() ActiveVersionResolver { return &activeVersionResolver{r} }

type activeVersionResolver struct{ *Resolver }
