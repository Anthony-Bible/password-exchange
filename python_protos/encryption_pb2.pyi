from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class DecryptedMessageRequest(_message.Message):
    __slots__ = ["Ciphertext", "key"]
    CIPHERTEXT_FIELD_NUMBER: _ClassVar[int]
    Ciphertext: _containers.RepeatedScalarFieldContainer[str]
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: bytes
    def __init__(self, Ciphertext: _Optional[_Iterable[str]] = ..., key: _Optional[bytes] = ...) -> None: ...

class DecryptedMessageResponse(_message.Message):
    __slots__ = ["plaintext"]
    PLAINTEXT_FIELD_NUMBER: _ClassVar[int]
    plaintext: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, plaintext: _Optional[_Iterable[str]] = ...) -> None: ...

class EncryptedMessageRequest(_message.Message):
    __slots__ = ["Key", "PlainText"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    Key: bytes
    PLAINTEXT_FIELD_NUMBER: _ClassVar[int]
    PlainText: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, PlainText: _Optional[_Iterable[str]] = ..., Key: _Optional[bytes] = ...) -> None: ...

class EncryptedMessageResponse(_message.Message):
    __slots__ = ["Ciphertext"]
    CIPHERTEXT_FIELD_NUMBER: _ClassVar[int]
    Ciphertext: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, Ciphertext: _Optional[_Iterable[str]] = ...) -> None: ...

class Randomrequest(_message.Message):
    __slots__ = ["randomLength"]
    RANDOMLENGTH_FIELD_NUMBER: _ClassVar[int]
    randomLength: int
    def __init__(self, randomLength: _Optional[int] = ...) -> None: ...

class Randomresponse(_message.Message):
    __slots__ = ["encryptionString", "encryptionbytes"]
    ENCRYPTIONBYTES_FIELD_NUMBER: _ClassVar[int]
    ENCRYPTIONSTRING_FIELD_NUMBER: _ClassVar[int]
    encryptionString: str
    encryptionbytes: bytes
    def __init__(self, encryptionbytes: _Optional[bytes] = ..., encryptionString: _Optional[str] = ...) -> None: ...
