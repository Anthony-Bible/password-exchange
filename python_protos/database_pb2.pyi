from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class InsertRequest(_message.Message):
    __slots__ = ["content", "max_view_count", "passphrase", "uuid"]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    MAX_VIEW_COUNT_FIELD_NUMBER: _ClassVar[int]
    PASSPHRASE_FIELD_NUMBER: _ClassVar[int]
    UUID_FIELD_NUMBER: _ClassVar[int]
    content: str
    max_view_count: int
    passphrase: str
    uuid: str
    def __init__(self, uuid: _Optional[str] = ..., content: _Optional[str] = ..., passphrase: _Optional[str] = ..., max_view_count: _Optional[int] = ...) -> None: ...

class SelectRequest(_message.Message):
    __slots__ = ["uuid"]
    UUID_FIELD_NUMBER: _ClassVar[int]
    uuid: str
    def __init__(self, uuid: _Optional[str] = ...) -> None: ...

class SelectResponse(_message.Message):
    __slots__ = ["content", "max_view_count", "passphrase", "uuid", "view_count"]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    MAX_VIEW_COUNT_FIELD_NUMBER: _ClassVar[int]
    PASSPHRASE_FIELD_NUMBER: _ClassVar[int]
    UUID_FIELD_NUMBER: _ClassVar[int]
    VIEW_COUNT_FIELD_NUMBER: _ClassVar[int]
    content: str
    max_view_count: int
    passphrase: str
    uuid: str
    view_count: int
    def __init__(self, uuid: _Optional[str] = ..., content: _Optional[str] = ..., passphrase: _Optional[str] = ..., view_count: _Optional[int] = ..., max_view_count: _Optional[int] = ...) -> None: ...
