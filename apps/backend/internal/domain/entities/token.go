package entities

type TokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	JTI       string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type Token struct {
	SignedToken string
	Claims      TokenClaims
}

type TokenPair struct {
	AccessToken  Token
	RefreshToken Token
}
