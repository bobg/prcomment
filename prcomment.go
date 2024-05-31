package prcomment

import (
	"context"

	"github.com/bobg/errors"
	"github.com/google/go-github/v62/github"
)

// Commenter is an object whose method AddOrUpdate adds a comment to a GitHub pull request
// or optionally updates an existing one.
type Commenter struct {
	// IsComment, if non-nil, is a function that returns true if a given comment is the one to update.
	IsComment func(*github.IssueComment) bool

	body   func(context.Context, *github.PullRequest) (string, error)
	prs    prsIntf
	issues issuesIntf
}

// NewCommenter creates a new Commenter object.
// The body function is called to generate the new or updated comment body from a given pull request.
func NewCommenter(client *github.Client, body func(context.Context, *github.PullRequest) (string, error)) *Commenter {
	return &Commenter{
		body:   body,
		prs:    client.PullRequests,
		issues: client.Issues,
	}
}

type prsIntf interface {
	Get(ctx context.Context, owner, reponame string, number int) (*github.PullRequest, *github.Response, error)
}

type issuesIntf interface {
	CreateComment(ctx context.Context, owner, reponame string, num int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	EditComment(ctx context.Context, owner, reponame string, commentID int64, newComment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	ListComments(ctx context.Context, owner, reponame string, number int, opts *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error)
}

func (c Commenter) AddOrUpdate(ctx context.Context, owner, reponame string, prnum int) error {
	pr, _, err := c.prs.Get(ctx, owner, reponame, prnum)
	if err != nil {
		return errors.Wrap(err, "getting pull request")
	}

	body, err := c.body(ctx, pr)
	if err != nil {
		return errors.Wrap(err, "getting comment body")
	}
	issueComment := &github.IssueComment{Body: &body}

	comments, _, err := c.issues.ListComments(ctx, owner, reponame, prnum, nil)
	if err != nil {
		return errors.Wrap(err, "listing PR comments")
	}

	if c.IsComment != nil {
		for _, comment := range comments {
			if c.IsComment(comment) {
				_, _, err = c.issues.EditComment(ctx, owner, reponame, *comment.ID, issueComment)
				return errors.Wrap(err, "updating PR comment")
			}
		}
	}

	_, _, err = c.issues.CreateComment(ctx, owner, reponame, prnum, issueComment)
	return errors.Wrap(err, "adding PR comment")
}
