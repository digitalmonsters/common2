package common

type DeviceType string

const (
	DeviceTypeIos     = DeviceType("ios")
	DeviceTypeAndroid = DeviceType("android")
	DeviceTypeWeb     = DeviceType("web")
)

type VerifiedByType string

const (
	VerifiedByTypeUnknown         = VerifiedByType("unknown")
	VerifiedByTypeManual          = VerifiedByType("manual")
	VerifiedByTypeCode            = VerifiedByType("code")
	VerifiedByTypeCde             = VerifiedByType("cde")
	VerifiedByTypeSharedContent   = VerifiedByType("shared_content")
	VerifiedByTypeUrlLink         = VerifiedByType("url_link")
	VerifiedByTypeUploadedContent = VerifiedByType("uploaded_content")
	VerifiedByTypeEmailInvites    = VerifiedByType("email_invites")
	VerifiedByTypeOther           = VerifiedByType("other")
	VerifiedByTypeNone            = VerifiedByType("")
)
