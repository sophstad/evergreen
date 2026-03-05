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
// This is temporary until we can remove the link from GitTag to APIGitTag
func (r *versionLiteResolver) GitTags(ctx context.Context, obj *model.Version) ([]*restModel.APIGitTag, error) {
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
func (r *versionLiteResolver) User(ctx context.Context, obj *model.Version) (*restModel.APIDBUser, error) {
	// userId, displayName, and emailAddress are always returned from the version document.
	// Other fields require a database call.
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
	// This is most likely a reaped user, so just return their info from version
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

// VersionLite returns VersionLiteResolver implementation.
func (r *Resolver) VersionLite() VersionLiteResolver { return &versionLiteResolver{r} }

type versionLiteResolver struct{ *Resolver }
