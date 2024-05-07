package models

type UserErr struct {
	userError string
}

func NewErr(userError string) UserErr {
	return UserErr{userError: userError}
}

func (e UserErr) Error() string {
	return e.userError
}
