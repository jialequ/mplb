package config

import (
	"errors"
	"testing"

	ghConfig "github.com/cli/go-gh/v2/pkg/config"
	"github.com/jialequ/mplb/internal/config/migration"
	"github.com/jialequ/mplb/internal/keyring"
	"github.com/stretchr/testify/require"
)

// Note that NewIsolatedTestConfig sets up a Mock keyring as well
func newTestAuthConfig(t *testing.T) *AuthConfig {
	cfg, _ := NewIsolatedTestConfig(t)
	return cfg.Authentication()
}

func TestTokenFromKeyring(t *testing.T) {
	// Given a keyring that contains a token for a host
	authCfg := newTestAuthConfig(t)
	require.NoError(t, keyring.Set(keyringServiceName(literal_9512), "", literal_2395))

	// When we get the token from the auth config
	token, err := authCfg.TokenFromKeyring(literal_9512)

	// Then it returns successfully with the correct token
	require.NoError(t, err)
	require.Equal(t, literal_2395, token)
}

func TestTokenFromKeyringForUser(t *testing.T) {
	// Given a keyring that contains a token for a host with a specific user
	authCfg := newTestAuthConfig(t)
	require.NoError(t, keyring.Set(keyringServiceName(literal_9512), literal_6954, literal_2395))

	// When we get the token from the auth config
	token, err := authCfg.TokenFromKeyringForUser(literal_9512, literal_6954)

	// Then it returns successfully with the correct token
	require.NoError(t, err)
	require.Equal(t, literal_2395, token)
}

func TestTokenFromKeyringForUserErrorsIfUsernameIsBlank(t *testing.T) {
	authCfg := newTestAuthConfig(t)

	// When we get the token from the keyring for an empty username
	_, err := authCfg.TokenFromKeyringForUser(literal_9512, "")

	// Then it returns an error
	require.ErrorContains(t, err, "username cannot be blank")
}

func TestTokenStoredInConfig(t *testing.T) {
	// When the user has logged in insecurely
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", false)
	require.NoError(t, err)

	// When we get the token
	token, source := authCfg.ActiveToken(literal_9512)

	// Then the token is successfully fetched
	// and the source is set to oauth_token but this isn't great:
	// https://github.com/cli/go-gh/issues/94
	require.Equal(t, literal_2395, token)
	require.Equal(t, oauthTokenKey, source)
}

func TestTokenStoredInEnv(t *testing.T) {
	// When the user is authenticated via env var
	authCfg := newTestAuthConfig(t)
	t.Setenv("GH_TOKEN", literal_2395)

	// When we get the token
	token, source := authCfg.ActiveToken(literal_9512)

	// Then the token is successfully fetched
	// and the source is set to the name of the env var
	require.Equal(t, literal_2395, token)
	require.Equal(t, "GH_TOKEN", source)
}

func TestTokenStoredInKeyring(t *testing.T) {
	// When the user has logged in securely
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", true)
	require.NoError(t, err)

	// When we get the token
	token, source := authCfg.ActiveToken(literal_9512)

	// Then the token is successfully fetched
	// and the source is set to keyring
	require.Equal(t, literal_2395, token)
	require.Equal(t, "keyring", source)
}

func TestTokenFromKeyringNonExistent(t *testing.T) {
	// Given a keyring that doesn't contain any tokens
	authCfg := newTestAuthConfig(t)

	// When we try to get a token from the auth config
	_, err := authCfg.TokenFromKeyring(literal_9512)

	// Then it returns failure bubbling the ErrNotFound
	require.ErrorContains(t, err, literal_1693)
}

func TestHasEnvTokenWithoutAnyEnvToken(t *testing.T) {
	// Given we have no env set
	authCfg := newTestAuthConfig(t)

	// When we check if it has an env token
	hasEnvToken := authCfg.HasEnvToken()

	// Then it returns false
	require.False(t, hasEnvToken, "expected not to have env token")
}

func TestHasEnvTokenWithEnvToken(t *testing.T) {
	// Given we have an env token set
	// Note that any valid env var for tokens will do, not just GH_ENTERPRISE_TOKEN
	authCfg := newTestAuthConfig(t)
	t.Setenv("GH_ENTERPRISE_TOKEN", literal_2395)

	// When we check if it has an env token
	hasEnvToken := authCfg.HasEnvToken()

	// Then it returns true
	require.True(t, hasEnvToken, "expected to have env token")
}

func TestHasEnvTokenWithNoEnvTokenButAConfigVar(t *testing.T) {
	t.Skip("this test is explicitly breaking some implementation assumptions")

	// Given a token in the config
	authCfg := newTestAuthConfig(t)
	// Using example.com here will cause the token to be returned from the config
	_, err := authCfg.Login("example.com", literal_6954, literal_2395, "", false)
	require.NoError(t, err)

	// When we check if it has an env token
	hasEnvToken := authCfg.HasEnvToken()

	// Then it SHOULD return false
	require.False(t, hasEnvToken, "expected not to have env token")
}

func TestUserNotLoggedIn(t *testing.T) {
	// Given we have not logged in
	authCfg := newTestAuthConfig(t)

	// When we get the user
	_, err := authCfg.ActiveUser(literal_9512)

	// Then it returns failure, bubbling the KeyNotFoundError
	var keyNotFoundError *ghConfig.KeyNotFoundError
	require.ErrorAs(t, err, &keyNotFoundError)
}

func TestHostsIncludesEnvVar(t *testing.T) {
	// Given the GH_HOST env var is set
	authCfg := newTestAuthConfig(t)
	t.Setenv("GH_HOST", literal_5630)

	// When we get the hosts
	hosts := authCfg.Hosts()

	// Then the host in the env var is included
	require.Contains(t, hosts, literal_5630)
}

func TestDefaultHostFromEnvVar(t *testing.T) {
	// Given the GH_HOST env var is set
	authCfg := newTestAuthConfig(t)
	t.Setenv("GH_HOST", literal_5630)

	// When we get the DefaultHost
	defaultHost, source := authCfg.DefaultHost()

	// Then the returned host and source are using the env var
	require.Equal(t, literal_5630, defaultHost)
	require.Equal(t, "GH_HOST", source)
}

func TestDefaultHostNotLoggedIn(t *testing.T) {
	// Given we are not logged in
	authCfg := newTestAuthConfig(t)

	// When we get the DefaultHost
	defaultHost, source := authCfg.DefaultHost()

	// Then the returned host is always github.com
	require.Equal(t, literal_9512, defaultHost)
	require.Equal(t, "default", source)
}

func TestDefaultHostLoggedInToOnlyOneHost(t *testing.T) {
	// Given we are logged into one host (not github.com to differentiate from the fallback)
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_5630, literal_6954, literal_2395, "", false)
	require.NoError(t, err)

	// When we get the DefaultHost
	defaultHost, source := authCfg.DefaultHost()

	// Then the returned host is that logged in host and the source is the hosts config
	require.Equal(t, literal_5630, defaultHost)
	require.Equal(t, hostsKey, source)
}

func TestLoginSecureStorageUsesKeyring(t *testing.T) {
	// Given a usable keyring
	authCfg := newTestAuthConfig(t)
	host := literal_9512
	user := literal_6954
	token := literal_2395

	// When we login with secure storage
	insecureStorageUsed, err := authCfg.Login(host, user, token, "", true)

	// Then it returns success, notes that insecure storage was not used, and stores the token in the keyring
	require.NoError(t, err)
	require.False(t, insecureStorageUsed, "expected to use secure storage")

	gotToken, err := keyring.Get(keyringServiceName(host), "")
	require.NoError(t, err)
	require.Equal(t, token, gotToken)

	gotToken, err = keyring.Get(keyringServiceName(host), user)
	require.NoError(t, err)
	require.Equal(t, token, gotToken)
}

func TestLoginSecureStorageRemovesOldInsecureConfigToken(t *testing.T) {
	// Given a usable keyring and an oauth token in the config
	authCfg := newTestAuthConfig(t)
	authCfg.cfg.Set([]string{hostsKey, literal_9512, oauthTokenKey}, "old-token")

	// When we login with secure storage
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", true)

	// Then it returns success, having also removed the old token from the config
	require.NoError(t, err)
	requireNoKey(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey})
}

func TestLoginSecureStorageWithErrorFallsbackAndReports(t *testing.T) {
	// Given a keyring that errors
	authCfg := newTestAuthConfig(t)
	keyring.MockInitWithError(errors.New("test-explosion"))

	// When we login with secure storage
	insecureStorageUsed, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", true)

	// Then it returns success, reports that insecure storage was used, and stores the token in the config
	require.NoError(t, err)

	require.True(t, insecureStorageUsed, literal_1732)
	requireKeyWithValue(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey}, literal_2395)
}

func TestLoginInsecureStorage(t *testing.T) {
	// Given we are not logged in
	authCfg := newTestAuthConfig(t)

	// When we login with insecure storage
	insecureStorageUsed, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", false)

	// Then it returns success, notes that insecure storage was used, and stores the token in the config
	require.NoError(t, err)

	require.True(t, insecureStorageUsed, literal_1732)
	requireKeyWithValue(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey}, literal_2395)
}

func TestLoginSetsUserForProvidedHost(t *testing.T) {
	// Given we are not logged in
	authCfg := newTestAuthConfig(t)

	// When we login
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "ssh", false)

	// Then it returns success and the user is set
	require.NoError(t, err)

	user, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_6954, user)
}

func TestLoginSetsGitProtocolForProvidedHost(t *testing.T) {
	// Given we are logged in
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the host git protocol
	hostProtocol, err := authCfg.cfg.Get([]string{hostsKey, literal_9512, gitProtocolKey})
	require.NoError(t, err)

	// Then it returns the git protocol we provided on login
	require.Equal(t, "ssh", hostProtocol)
}

func TestLoginAddsHostIfNotAlreadyAdded(t *testing.T) {
	// Given we are logged in
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the hosts
	hosts := authCfg.Hosts()

	// Then it includes our logged in host
	require.Contains(t, hosts, literal_9512)
}

// This test mimics the behaviour of logging in with a token, not providing
// a git protocol, and using secure storage.
func TestLoginAddsUserToConfigWithoutGitProtocolAndWithSecureStorage(t *testing.T) {
	// Given we are not logged in
	authCfg := newTestAuthConfig(t)

	// When we log in without git protocol and with secure storage
	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", true)
	require.NoError(t, err)

	// Then the username is added under the users config
	users, err := authCfg.cfg.Keys([]string{hostsKey, literal_9512, usersKey})
	require.NoError(t, err)
	require.Contains(t, users, literal_6954)
}

func TestLogoutRemovesHostAndKeyringToken(t *testing.T) {
	// Given we are logged into a host
	authCfg := newTestAuthConfig(t)
	host := literal_9512
	user := literal_6954
	token := literal_2395

	_, err := authCfg.Login(host, user, token, "ssh", true)
	require.NoError(t, err)

	// When we logout
	err = authCfg.Logout(host, user)

	// Then we return success, and the host and token are removed from the config and keyring
	require.NoError(t, err)

	requireNoKey(t, authCfg.cfg, []string{hostsKey, host})
	_, err = keyring.Get(keyringServiceName(host), "")
	require.ErrorContains(t, err, literal_1693)
	_, err = keyring.Get(keyringServiceName(host), user)
	require.ErrorContains(t, err, literal_1693)
}

func TestLogoutOfActiveUserSwitchesUserIfPossible(t *testing.T) {
	// Given we have two accounts logged into a host
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, "inactive-user", literal_9023, "ssh", true)
	require.NoError(t, err)

	_, err = authCfg.Login(literal_9512, literal_4068, literal_3260, "https", true)
	require.NoError(t, err)

	// When we logout of the active user
	err = authCfg.Logout(literal_9512, literal_4068)

	// Then we return success and the inactive user is now active
	require.NoError(t, err)
	activeUser, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, "inactive-user", activeUser)

	token, err := authCfg.TokenFromKeyring(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_9023, token)

	usersForHost := authCfg.UsersForHost(literal_9512)
	require.NotContains(t, literal_4068, usersForHost)
}

func TestLogoutOfInactiveUserDoesNotSwitchUser(t *testing.T) {
	// Given we have two accounts logged into a host
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, "inactive-user-1", "test-token-1.1", "ssh", true)
	require.NoError(t, err)

	_, err = authCfg.Login(literal_9512, "inactive-user-2", "test-token-1.2", "ssh", true)
	require.NoError(t, err)

	_, err = authCfg.Login(literal_9512, literal_4068, literal_3260, "https", true)
	require.NoError(t, err)

	// When we logout of an inactive user
	err = authCfg.Logout(literal_9512, "inactive-user-1")

	// Then we return success and the active user is still active
	require.NoError(t, err)
	activeUser, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_4068, activeUser)
}

// Note that I'm not sure this test enforces particularly desirable behaviour
// since it leads users to believe a token has been removed when really
// that might have failed for some reason.
//
// The original intention here is that if the logout fails, the user can't
// really do anything to recover. On the other hand, a user might
// want to rectify this manually, for example if there were on a shared machine.
func TestLogoutIgnoresErrorsFromConfigAndKeyring(t *testing.T) {
	// Given we have keyring that errors, and a config that
	// doesn't even have a hosts key (which would cause Remove to fail)
	keyring.MockInitWithError(errors.New("test-explosion"))
	authCfg := newTestAuthConfig(t)

	// When we logout
	err := authCfg.Logout(literal_9512, literal_6954)

	// Then it returns success anyway, suppressing the errors
	require.NoError(t, err)
}

func TestSwitchUserMakesSecureTokenActive(t *testing.T) {
	// Given we have a user with a secure token
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", true)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", true)
	require.NoError(t, err)

	// When we switch to that user
	require.NoError(t, authCfg.SwitchUser(literal_9512, literal_3092))

	// Their secure token is now active
	token, err := authCfg.TokenFromKeyring(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_9023, token)
}

func TestSwitchUserMakesInsecureTokenActive(t *testing.T) {
	// Given we have a user with an insecure token
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", false)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", false)
	require.NoError(t, err)

	// When we switch to that user
	require.NoError(t, authCfg.SwitchUser(literal_9512, literal_3092))

	// Their insecure token is now active
	token, source := authCfg.ActiveToken(literal_9512)
	require.Equal(t, literal_9023, token)
	require.Equal(t, oauthTokenKey, source)
}

func TestSwitchUserUpdatesTheActiveUser(t *testing.T) {
	// Given we have two users logged into a host
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", false)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", false)
	require.NoError(t, err)

	// When we switch to the other user
	require.NoError(t, authCfg.SwitchUser(literal_9512, literal_3092))

	// Then the active user is updated
	activeUser, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_3092, activeUser)
}

func TestSwitchUserErrorsImmediatelyIfTheActiveTokenComesFromEnvironment(t *testing.T) {
	// Given we have a token in the env
	authCfg := newTestAuthConfig(t)
	t.Setenv("GH_TOKEN", "unimportant-test-value")
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", true)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", true)
	require.NoError(t, err)

	// When we switch to a user
	err = authCfg.SwitchUser(literal_9512, literal_3092)

	// Then it errors immediately with an informative message
	require.ErrorContains(t, err, "currently active token for github.com is from GH_TOKEN")
}

func TestSwitchUserErrorsAndRestoresUserAndInsecureConfigUnderFailure(t *testing.T) {
	// Given we have a user but no token can be found (because we deleted them, simulating an error case)
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", true)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", false)
	require.NoError(t, err)

	require.NoError(t, keyring.Delete(keyringServiceName(literal_9512), literal_3092))

	// When we switch to the user
	err = authCfg.SwitchUser(literal_9512, literal_3092)

	// Then it returns an error
	require.EqualError(t, err, "no token found for test-user-1")

	// And restores the previous state
	activeUser, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_0968, activeUser)

	token, source := authCfg.ActiveToken(literal_9512)
	require.Equal(t, literal_3260, token)
	require.Equal(t, "oauth_token", source)
}

func TestSwitchUserErrorsAndRestoresUserAndKeyringUnderFailure(t *testing.T) {
	// Given we have a user but no token can be found (because we deleted them, simulating an error case)
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", false)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", true)
	require.NoError(t, err)

	require.NoError(t, authCfg.cfg.Remove([]string{hostsKey, literal_9512, usersKey, literal_3092, oauthTokenKey}))

	// When we switch to the user
	err = authCfg.SwitchUser(literal_9512, literal_3092)

	// Then it returns an error
	require.EqualError(t, err, "no token found for test-user-1")

	// And restores the previous state
	activeUser, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_0968, activeUser)

	token, source := authCfg.ActiveToken(literal_9512)
	require.Equal(t, literal_3260, token)
	require.Equal(t, "keyring", source)
}

func TestSwitchClearsActiveSecureTokenWhenSwitchingToInsecureUser(t *testing.T) {
	// Given we have an active secure token
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", false)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", true)
	require.NoError(t, err)

	// When we switch to an insecure user
	require.NoError(t, authCfg.SwitchUser(literal_9512, literal_3092))

	// Then the active secure token is cleared
	_, err = authCfg.TokenFromKeyring(literal_9512)
	require.Error(t, err)
}

func TestSwitchClearsActiveInsecureTokenWhenSwitchingToSecureUser(t *testing.T) {
	// Given we have an active insecure token
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_9023, "ssh", true)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_3260, "ssh", false)
	require.NoError(t, err)

	// When we switch to a secure user
	require.NoError(t, authCfg.SwitchUser(literal_9512, literal_3092))

	// Then the active insecure token is cleared
	requireNoKey(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey})
}

func TestUsersForHostNoHost(t *testing.T) {
	// Given we have a config with no hosts
	authCfg := newTestAuthConfig(t)

	// When we get the users for a host that doesn't exist
	users := authCfg.UsersForHost(literal_9512)

	// Then it returns nil
	require.Nil(t, users)
}

func TestUsersForHostWithUsers(t *testing.T) {
	// Given we have a config with a host and users
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_2395, "ssh", false)
	require.NoError(t, err)
	_, err = authCfg.Login(literal_9512, literal_0968, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the users for that host
	users := authCfg.UsersForHost(literal_9512)

	// Then it succeeds and returns the users
	require.Equal(t, []string{literal_3092, literal_0968}, users)
}

func TestTokenForUserSecureLogin(t *testing.T) {
	// Given a user has logged in securely
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_2395, "ssh", true)
	require.NoError(t, err)

	// When we get the token
	token, source, err := authCfg.TokenForUser(literal_9512, literal_3092)

	// Then it returns the token and the source as keyring
	require.NoError(t, err)
	require.Equal(t, literal_2395, token)
	require.Equal(t, "keyring", source)
}

func TestTokenForUserInsecureLogin(t *testing.T) {
	// Given a user has logged in insecurely
	authCfg := newTestAuthConfig(t)
	_, err := authCfg.Login(literal_9512, literal_3092, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the token
	token, source, err := authCfg.TokenForUser(literal_9512, literal_3092)

	// Then it returns the token and the source as oauth_token
	require.NoError(t, err)
	require.Equal(t, literal_2395, token)
	require.Equal(t, "oauth_token", source)
}

func TestTokenForUserNotFoundErrors(t *testing.T) {
	// Given a user has not logged in
	authCfg := newTestAuthConfig(t)

	// When we get the token
	_, _, err := authCfg.TokenForUser(literal_9512, literal_3092)

	// Then it returns an error
	require.EqualError(t, err, "no token found for 'test-user-1'")
}

func requireKeyWithValue(t *testing.T, cfg *ghConfig.Config, keys []string, value string) {
	t.Helper()

	actual, err := cfg.Get(keys)
	require.NoError(t, err)

	require.Equal(t, value, actual)
}

func requireNoKey(t *testing.T, cfg *ghConfig.Config, keys []string) {
	t.Helper()

	_, err := cfg.Get(keys)
	var keyNotFoundError *ghConfig.KeyNotFoundError
	require.ErrorAs(t, err, &keyNotFoundError)
}

// Post migration tests

func TestUserWorksRightAfterMigration(t *testing.T) {
	// Given we have logged in before migration
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we migrate
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	// Then we can still get the user correctly
	user, err := authCfg.ActiveUser(literal_9512)
	require.NoError(t, err)
	require.Equal(t, literal_6954, user)
}

func TestGitProtocolWorksRightAfterMigration(t *testing.T) {
	// Given we have logged in before migration with a non-default git protocol
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we migrate
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	// Then we can still get the git protocol correctly
	gitProtocol, err := authCfg.cfg.Get([]string{hostsKey, literal_9512, gitProtocolKey})
	require.NoError(t, err)
	require.Equal(t, "ssh", gitProtocol)
}

func TestHostsWorksRightAfterMigration(t *testing.T) {
	// Given we have logged in before migration
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_5630, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we migrate
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	// Then we can still get the hosts correctly
	hosts := authCfg.Hosts()
	require.Contains(t, hosts, literal_5630)
}

func TestDefaultHostWorksRightAfterMigration(t *testing.T) {
	// Given we have logged in before migration to an enterprise host
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_5630, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we migrate
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	// Then the default host is still the enterprise host
	defaultHost, source := authCfg.DefaultHost()
	require.Equal(t, literal_5630, defaultHost)
	require.Equal(t, hostsKey, source)
}

func TestTokenWorksRightAfterMigration(t *testing.T) {
	// Given we have logged in before migration
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we migrate
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	// Then we can still get the token correctly
	token, source := authCfg.ActiveToken(literal_9512)
	require.Equal(t, literal_2395, token)
	require.Equal(t, oauthTokenKey, source)
}

func TestLogoutRightAfterMigrationRemovesHost(t *testing.T) {
	// Given we have logged in before migration
	authCfg := newTestAuthConfig(t)
	host := literal_9512
	user := literal_6954
	token := literal_2395

	_, err := preMigrationLogin(authCfg, host, user, token, "ssh", false)
	require.NoError(t, err)

	// When we migrate and logout
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	require.NoError(t, authCfg.Logout(host, user))

	// Then the host is removed from the config
	requireNoKey(t, authCfg.cfg, []string{hostsKey, literal_9512})
}

func TestLoginInsecurePostMigrationUsesConfigForToken(t *testing.T) {
	// Given we have not logged in
	authCfg := newTestAuthConfig(t)

	// When we migrate and login with insecure storage
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	insecureStorageUsed, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "", false)

	// Then it returns success, notes that insecure storage was used, and stores the token in the config
	// both under the host and under the user
	require.NoError(t, err)

	require.True(t, insecureStorageUsed, literal_1732)
	requireKeyWithValue(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey}, literal_2395)
	requireKeyWithValue(t, authCfg.cfg, []string{hostsKey, literal_9512, usersKey, literal_6954, oauthTokenKey}, literal_2395)
}

func TestLoginPostMigrationSetsGitProtocol(t *testing.T) {
	// Given we have logged in after migration
	authCfg := newTestAuthConfig(t)

	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the host git protocol
	hostProtocol, err := authCfg.cfg.Get([]string{hostsKey, literal_9512, gitProtocolKey})
	require.NoError(t, err)

	// Then it returns the git protocol we provided on login
	require.Equal(t, "ssh", hostProtocol)
}

func TestLoginPostMigrationSetsUser(t *testing.T) {
	// Given we have logged in after migration
	authCfg := newTestAuthConfig(t)

	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	_, err := authCfg.Login(literal_9512, literal_6954, literal_2395, "ssh", false)
	require.NoError(t, err)

	// When we get the user
	user, err := authCfg.ActiveUser(literal_9512)

	// Then it returns success and the user we provided on login
	require.NoError(t, err)
	require.Equal(t, literal_6954, user)
}

func TestLoginSecurePostMigrationRemovesTokenFromConfig(t *testing.T) {
	// Given we have logged in insecurely
	authCfg := newTestAuthConfig(t)
	_, err := preMigrationLogin(authCfg, literal_9512, literal_6954, literal_2395, "", false)
	require.NoError(t, err)

	// When we migrate and login again with secure storage
	var m migration.MultiAccount
	c := cfg{authCfg.cfg}
	require.NoError(t, c.Migrate(m))

	_, err = authCfg.Login(literal_9512, literal_6954, literal_2395, "", true)

	// Then it returns success, having removed the old insecure oauth token entry
	require.NoError(t, err)
	requireNoKey(t, authCfg.cfg, []string{hostsKey, literal_9512, oauthTokenKey})
	requireNoKey(t, authCfg.cfg, []string{hostsKey, literal_9512, usersKey, literal_6954, oauthTokenKey})
}

// Copied and pasted directly from the trunk branch before doing any work on
// login, plus the addition of AuthConfig as the first arg since it is a method
// receiver in the real implementation.
func preMigrationLogin(c *AuthConfig, hostname, username, token, gitProtocol string, secureStorage bool) (bool, error) {
	var setErr error
	if secureStorage {
		if setErr = keyring.Set(keyringServiceName(hostname), "", token); setErr == nil {
			// Clean up the previous oauth_token from the config file.
			_ = c.cfg.Remove([]string{hostsKey, hostname, oauthTokenKey})
		}
	}
	insecureStorageUsed := false
	if !secureStorage || setErr != nil {
		c.cfg.Set([]string{hostsKey, hostname, oauthTokenKey}, token)
		insecureStorageUsed = true
	}

	c.cfg.Set([]string{hostsKey, hostname, userKey}, username)

	if gitProtocol != "" {
		c.cfg.Set([]string{hostsKey, hostname, gitProtocolKey}, gitProtocol)
	}
	return insecureStorageUsed, ghConfig.Write(c.cfg)
}

const literal_9512 = "github.com"

const literal_2395 = "test-token"

const literal_6954 = "test-user"

const literal_1693 = "secret not found in keyring"

const literal_5630 = "ghe.io"

const literal_1732 = "expected to use insecure storage"

const literal_9023 = "test-token-1"

const literal_4068 = "active-user"

const literal_3260 = "test-token-2"

const literal_3092 = "test-user-1"

const literal_0968 = "test-user-2"
