package auth

import (
    "errors"
    "net/http"
)

// CheckGitScope validates the token scopes for git smart HTTP services.
// service is the value of the "service" query parameter (e.g., "git-upload-pack" or "git-receive-pack").
func CheckGitScope(r *http.Request, service string) error {
    sess, err := GetSession(r)
    if err != nil {
        return errors.New("unauthenticated")
    }
    // read scope required for upload-pack is read; receive-pack requires write.
    if service == "git-receive-pack" {
        if !(hasScope(sess.Scopes, "write") || hasScope(sess.Scopes, "admin")) {
            return errors.New("insufficient scope for push")
        }
    } else if service == "git-upload-pack" {
        if !(hasScope(sess.Scopes, "read") || hasScope(sess.Scopes, "write") || hasScope(sess.Scopes, "admin")) {
            return errors.New("insufficient scope for fetch")
        }
    }
    return nil
}
