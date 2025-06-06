# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: database.proto
# Protobuf Python Version: 5.29.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    0,
    '',
    'database.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0e\x64\x61tabase.proto\x12\ndatabasepb\x1a\x1bgoogle/protobuf/empty.proto\"\x1d\n\rSelectRequest\x12\x0c\n\x04uuid\x18\x01 \x01(\t\"o\n\x0eSelectResponse\x12\x0c\n\x04uuid\x18\x01 \x01(\t\x12\x0f\n\x07\x63ontent\x18\x02 \x01(\t\x12\x12\n\npassphrase\x18\x03 \x01(\t\x12\x12\n\nview_count\x18\x04 \x01(\x05\x12\x16\n\x0emax_view_count\x18\x05 \x01(\x05\"s\n\rInsertRequest\x12\x0c\n\x04uuid\x18\x01 \x01(\t\x12\x0f\n\x07\x63ontent\x18\x02 \x01(\t\x12\x12\n\npassphrase\x18\x03 \x01(\t\x12\x16\n\x0emax_view_count\x18\x04 \x01(\x05\x12\x17\n\x0frecipient_email\x18\x05 \x01(\t\"n\n\x1aGetUnviewedMessagesRequest\x12\x18\n\x10older_than_hours\x18\x01 \x01(\x05\x12\x15\n\rmax_reminders\x18\x02 \x01(\x05\x12\x1f\n\x17reminder_interval_hours\x18\x03 \x01(\x05\"t\n\x0fUnviewedMessage\x12\x12\n\nmessage_id\x18\x01 \x01(\x05\x12\x11\n\tunique_id\x18\x02 \x01(\t\x12\x17\n\x0frecipient_email\x18\x03 \x01(\t\x12\x0f\n\x07\x63reated\x18\x04 \x01(\t\x12\x10\n\x08\x64\x61ys_old\x18\x05 \x01(\x05\"L\n\x1bGetUnviewedMessagesResponse\x12-\n\x08messages\x18\x01 \x03(\x0b\x32\x1b.databasepb.UnviewedMessage\"?\n\x12LogReminderRequest\x12\x12\n\nmessage_id\x18\x01 \x01(\x05\x12\x15\n\remail_address\x18\x02 \x01(\t\"/\n\x19GetReminderHistoryRequest\x12\x12\n\nmessage_id\x18\x01 \x01(\x05\"q\n\x10ReminderLogEntry\x12\x12\n\nmessage_id\x18\x01 \x01(\x05\x12\x15\n\remail_address\x18\x02 \x01(\t\x12\x16\n\x0ereminder_count\x18\x03 \x01(\x05\x12\x1a\n\x12last_reminder_sent\x18\x04 \x01(\t\"K\n\x1aGetReminderHistoryResponse\x12-\n\x07\x65ntries\x18\x01 \x03(\x0b\x32\x1c.databasepb.ReminderLogEntry2\xfe\x03\n\tdbService\x12\x41\n\x06Select\x12\x19.databasepb.SelectRequest\x1a\x1a.databasepb.SelectResponse\"\x00\x12=\n\x06Insert\x12\x19.databasepb.InsertRequest\x1a\x16.google.protobuf.Empty\"\x00\x12\x45\n\nGetMessage\x12\x19.databasepb.SelectRequest\x1a\x1a.databasepb.SelectResponse\"\x00\x12t\n\x1fGetUnviewedMessagesForReminders\x12&.databasepb.GetUnviewedMessagesRequest\x1a\'.databasepb.GetUnviewedMessagesResponse\"\x00\x12K\n\x0fLogReminderSent\x12\x1e.databasepb.LogReminderRequest\x1a\x16.google.protobuf.Empty\"\x00\x12\x65\n\x12GetReminderHistory\x12%.databasepb.GetReminderHistoryRequest\x1a&.databasepb.GetReminderHistoryResponse\"\x00\x42;Z9github.com/Anthony-Bible/password-exchange/app/databasepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'database_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z9github.com/Anthony-Bible/password-exchange/app/databasepb'
  _globals['_SELECTREQUEST']._serialized_start=59
  _globals['_SELECTREQUEST']._serialized_end=88
  _globals['_SELECTRESPONSE']._serialized_start=90
  _globals['_SELECTRESPONSE']._serialized_end=201
  _globals['_INSERTREQUEST']._serialized_start=203
  _globals['_INSERTREQUEST']._serialized_end=318
  _globals['_GETUNVIEWEDMESSAGESREQUEST']._serialized_start=320
  _globals['_GETUNVIEWEDMESSAGESREQUEST']._serialized_end=430
  _globals['_UNVIEWEDMESSAGE']._serialized_start=432
  _globals['_UNVIEWEDMESSAGE']._serialized_end=548
  _globals['_GETUNVIEWEDMESSAGESRESPONSE']._serialized_start=550
  _globals['_GETUNVIEWEDMESSAGESRESPONSE']._serialized_end=626
  _globals['_LOGREMINDERREQUEST']._serialized_start=628
  _globals['_LOGREMINDERREQUEST']._serialized_end=691
  _globals['_GETREMINDERHISTORYREQUEST']._serialized_start=693
  _globals['_GETREMINDERHISTORYREQUEST']._serialized_end=740
  _globals['_REMINDERLOGENTRY']._serialized_start=742
  _globals['_REMINDERLOGENTRY']._serialized_end=855
  _globals['_GETREMINDERHISTORYRESPONSE']._serialized_start=857
  _globals['_GETREMINDERHISTORYRESPONSE']._serialized_end=932
  _globals['_DBSERVICE']._serialized_start=935
  _globals['_DBSERVICE']._serialized_end=1445
# @@protoc_insertion_point(module_scope)
