package auth

func hasScope(scopes []string, target string) bool {
    for _, s := range scopes {
        if s == target {
            return true
        }
    }
    return false
}
