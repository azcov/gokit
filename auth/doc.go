// Package auth defines authentication interfaces: the core Provider (login,
// token lifecycle), the OAuthProvider (social login), and optional capability
// interfaces (UserManager, PasswordManager, SessionManager).
//
// Implementations: auth/basic, auth/clerk, auth/supabase, auth/betterauth,
// auth/oauth/* (google, github, facebook, discord, microsoft, twitter),
// and auth/otp for passwordless one-time codes.
package auth
