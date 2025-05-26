from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Message(_message.Message):
    __slots__ = ["Captcha", "Content", "Errors", "Hidden", "OtherEmail", "OtherLastName", "Uniqueid", "Url", "email", "firstname", "otherfirstname"]
    CAPTCHA_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    Captcha: str
    Content: str
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    Errors: str
    FIRSTNAME_FIELD_NUMBER: _ClassVar[int]
    HIDDEN_FIELD_NUMBER: _ClassVar[int]
    Hidden: str
    OTHEREMAIL_FIELD_NUMBER: _ClassVar[int]
    OTHERFIRSTNAME_FIELD_NUMBER: _ClassVar[int]
    OTHERLASTNAME_FIELD_NUMBER: _ClassVar[int]
    OtherEmail: str
    OtherLastName: str
    UNIQUEID_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    Uniqueid: str
    Url: str
    email: str
    firstname: str
    otherfirstname: str
    def __init__(self, email: _Optional[str] = ..., firstname: _Optional[str] = ..., otherfirstname: _Optional[str] = ..., OtherLastName: _Optional[str] = ..., OtherEmail: _Optional[str] = ..., Uniqueid: _Optional[str] = ..., Content: _Optional[str] = ..., Errors: _Optional[str] = ..., Url: _Optional[str] = ..., Hidden: _Optional[str] = ..., Captcha: _Optional[str] = ...) -> None: ...
