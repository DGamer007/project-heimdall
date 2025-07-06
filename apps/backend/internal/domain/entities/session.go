package entities

type Session struct {
	Id             string
	UserId         string
	AccessTokenJTI string
	RefreshToken   string
	ExpiresAt      int64
}

type SessionRepository interface {
	CreateOne(session *Session) (string, error)
	DeleteOneByAccessTokenJTI(jti string) (bool, error)
	CheckIfExistsByIdentifier(userId string, accessTokenJTI string, refreshToken string) (bool, error)
	ReplaceSession(oldAccessTokenJTI string, newSession *Session) (string, error)
}
