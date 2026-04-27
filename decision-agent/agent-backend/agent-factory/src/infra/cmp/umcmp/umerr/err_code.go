package umerr

type ErrCode = int

const (
	UserNotFound       ErrCode = 400019001
	DepartmentNotFound ErrCode = 400019002
	GroupNotFound      ErrCode = 400019003

	ContactorNotFound ErrCode = 400019004
)
