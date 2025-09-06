package pkg

import "errors"

type TeleCommunicationErrorType error

type TeleCommunicationError struct {
	ErrorType TeleCommunicationErrorType
	ErrorCode int
}

var (
	TeleCommunicationDict = map[TeleCommunicationErrorType]TeleCommunicationError{
		ErrNoPermission: {
			ErrorType: ErrNoPermission,
			ErrorCode: 401,
		},
		ErrParamError: {
			ErrorType: ErrParamError,
			ErrorCode: 400,
		},
		ErrUserNotFound: {
			ErrorType: ErrUserNotFound,
			ErrorCode: 404,
		},
	}
)

func GetTeleCommunicationErrorCode(errorType TeleCommunicationErrorType) int {
	return TeleCommunicationDict[errorType].ErrorCode
}

var (
	// 没有权限
	ErrNoPermission TeleCommunicationErrorType = errors.New("没有权限")

	// 参数错误
	ErrParamError TeleCommunicationErrorType = errors.New("参数错误")

	// 用户不存在
	ErrUserNotFound TeleCommunicationErrorType = errors.New("用户不存在")

	// 无效的token
	ErrInvalidToken TeleCommunicationErrorType = errors.New("无效的token")
)
