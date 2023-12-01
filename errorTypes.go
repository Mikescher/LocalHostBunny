package bunny

import (
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
)

var (
	ErrInternal       = exerr.TypeInternal
	ErrPanic          = exerr.TypePanic
	ErrWrap           = exerr.TypeWrap
	ErrNotImplemented = exerr.TypeNotImplemented

	ErrBindFailURI      = exerr.TypeBindFailURI
	ErrBindFailQuery    = exerr.TypeBindFailQuery
	ErrBindFailJSON     = exerr.TypeBindFailJSON
	ErrBindFailFormData = exerr.TypeBindFailFormData

	ErrUnauthorized = exerr.TypeUnauthorized
	ErrAuthFailed   = exerr.TypeAuthFailed

	ErrDatabaseError     = exerr.NewType("DATABASE_ERROR", langext.Ptr(500))
	ErrFilesystemError   = exerr.NewType("FILESYSTEM_ERROR", langext.Ptr(500))
	ErrInvalidStateError = exerr.NewType("INV_STATE_ERROR", langext.Ptr(500))

	ErrInvalidRequestParams = exerr.NewType("INVALID_REQUEST_PARAMETER", langext.Ptr(400))
	ErrMissingRequestParams = exerr.NewType("MISSING_REQUEST_PARAMETER", langext.Ptr(400))
	ErrSelfDelete           = exerr.NewType("SELF_DELETE", langext.Ptr(400))
	ErrInvalidRefKey        = exerr.NewType("INVALID_REF_KEY", langext.Ptr(400))
	ErrInvalidMimeType      = exerr.NewType("INVALID_MIME_TYPE", langext.Ptr(400))
	ErrInvalidBlobType      = exerr.NewType("INVALID_BLOB_TYPE", langext.Ptr(400))
	ErrInvalidAuthType      = exerr.NewType("INVALID_AUTH_TYPE", langext.Ptr(400))
	ErrInvalidSubType       = exerr.NewType("INVALID_SUB_TYPE", langext.Ptr(400))
	ErrPostURLFormat        = exerr.NewType("POST_URL_FORMAT", langext.Ptr(400))

	ErrEntityNotFound     = exerr.NewType("ENTITY_NOT_FOUND", langext.Ptr(400))
	ErrInvalidCursorToken = exerr.NewType("INVALID_CURSOR_TOKEN", langext.Ptr(400))

	ErrUsernameCollision = exerr.NewType("USERNAME_COLLISION", langext.Ptr(400))
	ErrEmailCollision    = exerr.NewType("EMAIL_COLLISION", langext.Ptr(400))

	ErrPreconditionFailed = exerr.NewType("PRECONDITION_FAILED", langext.Ptr(400))

	ErrMissingPermissions = exerr.NewType("MISSING_PERMISSIONS", langext.Ptr(400))

	ErrWrongOldPassword = exerr.NewType("WRONG_OLD_PASSWORD", langext.Ptr(400))
	ErrWrongSecret      = exerr.NewType("WRONG_SECRET", langext.Ptr(400))

	ErrMarshalBSON = exerr.NewType("MARSHAL_BSON", langext.Ptr(400))

	ErrInvalidJWT = exerr.NewType("INVALID_JWT", langext.Ptr(400))

	ErrFAPIError       = exerr.NewType("FAPI_ERROR", langext.Ptr(500))
	ErrFAPIUnsupported = exerr.NewType("FAPI_UNSUPPORTED", langext.Ptr(500))
	ErrSolseitAPIError = exerr.NewType("SOLSEIT_API_ERROR", langext.Ptr(500))
	ErrScraper         = exerr.NewType("SCRAPER", langext.Ptr(500))

	ErrInvalidEnum = exerr.NewType("INVALID_ENUM", langext.Ptr(500))

	ErrJob = exerr.NewType("JOB", langext.Ptr(500))

	ErrUnmarshalJSON   = exerr.NewType("UNMARSHAL_JSON", langext.Ptr(500))
	ErrFeatureDisabled = exerr.NewType("FEATURE_DISABLED", langext.Ptr(500))
)
