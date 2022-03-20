package services

const NicknameMaxLen = 20

const StatusOK = 0
const InternalError = 900000
const DbConnErr = 900001
const WriteErr = 901000
const ReadErr = 901001
const NotFound = 902000
const PermissionDenied = 902001
const Unauthorized = 902002
const IllegalParams = 902003
const AlreadyExist = 902004

const InvalidNickname = 102000
const EmailUnavailable = 102001
const InvalidPassword = 102010
const UserNotExist = 102020

const InvalidPatchCommand = 103003

const TFANotEnabled = 202000
const Wrong2FACode = 202001
const Illegal2FASecret = 202002
const TFAAttemptLimited = 202003
const TFAAlreadyEnabled = 202004

const InvalidInviteCode = 201001
