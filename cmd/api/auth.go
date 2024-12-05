package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	Issuer      string // eg domain.com
	Audience    string // who is it intended for
	Secret      string
	TokenExpiry time.Duration

	// optional but common for custom expiry period for the token (longer than the default short period)
	RefreshExpiry time.Duration

	//we will give refresh token to the user as a http-only secure cookie (not accessibkle from JS)
	// which will be included swith all requests sent to our backend from the frontend
	CookieDomain string // eg domain.com
	CookiePath   string // eg '/'
	CookieName   string
}

// set minimum data required to issue a token
type jwtUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`  // the token we actually issue
	RefreshToken string `json:"refresh_token"` // to be refreshed with
}

// NOTES: when using JWTs, one of the data fields must be 'claims' in which u store info about the authenticated user.
//
//	it should have only one field (very litle info)
type Claims struct {
	// NOTES: 'jwt.RegisteredClaims' is from the 'jwt-go' package you would have installed already
	jwt.RegisteredClaims
}

// To start, we need to generate a token
func (j *Auth) GenerateTokenPair(user *jwtUser) (TokenPairs, error) {
	// Create a token (an empty token object)
	token := jwt.New(jwt.SigningMethodHS256)

	// set the claims (what does this token claim to be eg name, subject, issuer, audience, etc)
	//	there are a few ways to do this, but map claims is the easiest approach
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	// the rest of the claims to be used here MUST be in 3-lowercase pre-defined characters (no more, no less)
	claims["sub"] = fmt.Sprint(user.ID) // now the claims subject ('sub') will always be the ID of the user in our DB
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	// set the expiry for the JWT (as short period of time)
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// create a signed token (sign the token)
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create a refresh token and set claims (there'll be fewer claims in the refresh token)
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// set the expiry for the refresh token (it will be longer than the expiry of the JWT itself)
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// create a signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create TokenPairs (var) & populate with signed tokens
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	// return TokenPairs
	return tokenPairs, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:    j.CookieName,
		Path:    j.CookiePath,
		Value:   refreshToken,
		Expires: time.Now().Add(j.RefreshExpiry),
		MaxAge:  int(j.RefreshExpiry.Seconds()),
		/////SameSite: http.SameSiteStrictMode,
		/////SameSite: http.SameSiteNoneMode, // Allows cross-site cookie
		SameSite: http.SameSiteLaxMode, // Changed to Lax for development
		////Domain:   j.CookieDomain,
		HttpOnly: true, // JS will have no access to this cookie
		Secure:   true, // in dev this will be false coz we're on localhost, but make sure to set it to true in prod
	}
}

// GetExpiredRefreshCookie deletes the cookie created above in GetRefreshCookie() from the iser's browser.
// NOTES: How to delete cookies-u set to be expired. Set its max age to be -1, & its expires to be time.Unix(0, 0)
//
//	(or anything in the past)
func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:    j.CookieName,
		Path:    j.CookiePath,
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
		/////SameSite: http.SameSiteStrictMode,
		/////SameSite: http.SameSiteNoneMode, // Allows cross-site cookie
		SameSite: http.SameSiteLaxMode, // Changed to Lax for development
		////Domain:   j.CookieDomain, ///// (temporarilly removing this as the the cookie specification needs no 'Domain' & needs 'Secure' to be set to true)
		HttpOnly: true, // JS will have no access to this cookie
		Secure:   true, // in dev this will be false coz we're on localhost, but make sure to set it to true in prod
	}
}

// GetTokenFromHeaderAndVerify is a handy function that can be used allover our app to restrict user access to
// routes. A handy place to call it is in middleware
func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	// add a header to our response
	w.Header().Add("Vary", "Authorization")

	// get auth header
	authHeader := r.Header.Get("Authorization")

	// sanity check
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	// split the header on spaces-coz we expect to see the word 'Bearer' followed by space, followed by the JWT
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	// check if header includes the word 'Bearer' at the first index of the split header content
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
	}

	// at this point we have the word 'Bearer' in the header
	token := headerParts[1]

	// declare an empty claims into which we will read our claims
	claims := &Claims{}

	// parse the token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// validate the signing mechanism
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})
	// check for errors (this will include expired tokens)
	if err != nil {
		// does the error have thje prefix 'token is expired by'
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	// confirm that this token was actually issue by us
	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil
}
