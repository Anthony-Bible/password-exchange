# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

import database_pb2 as database__pb2
from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2

GRPC_GENERATED_VERSION = '1.71.0'
GRPC_VERSION = grpc.__version__
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    raise RuntimeError(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in database_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
    )


class dbServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Select = channel.unary_unary(
                '/databasepb.dbService/Select',
                request_serializer=database__pb2.SelectRequest.SerializeToString,
                response_deserializer=database__pb2.SelectResponse.FromString,
                _registered_method=True)
        self.Insert = channel.unary_unary(
                '/databasepb.dbService/Insert',
                request_serializer=database__pb2.InsertRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                _registered_method=True)
        self.GetMessage = channel.unary_unary(
                '/databasepb.dbService/GetMessage',
                request_serializer=database__pb2.SelectRequest.SerializeToString,
                response_deserializer=database__pb2.SelectResponse.FromString,
                _registered_method=True)
        self.GetUnviewedMessagesForReminders = channel.unary_unary(
                '/databasepb.dbService/GetUnviewedMessagesForReminders',
                request_serializer=database__pb2.GetUnviewedMessagesRequest.SerializeToString,
                response_deserializer=database__pb2.GetUnviewedMessagesResponse.FromString,
                _registered_method=True)
        self.LogReminderSent = channel.unary_unary(
                '/databasepb.dbService/LogReminderSent',
                request_serializer=database__pb2.LogReminderRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                _registered_method=True)
        self.GetReminderHistory = channel.unary_unary(
                '/databasepb.dbService/GetReminderHistory',
                request_serializer=database__pb2.GetReminderHistoryRequest.SerializeToString,
                response_deserializer=database__pb2.GetReminderHistoryResponse.FromString,
                _registered_method=True)


class dbServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Select(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Insert(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetMessage(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetUnviewedMessagesForReminders(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def LogReminderSent(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetReminderHistory(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_dbServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Select': grpc.unary_unary_rpc_method_handler(
                    servicer.Select,
                    request_deserializer=database__pb2.SelectRequest.FromString,
                    response_serializer=database__pb2.SelectResponse.SerializeToString,
            ),
            'Insert': grpc.unary_unary_rpc_method_handler(
                    servicer.Insert,
                    request_deserializer=database__pb2.InsertRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'GetMessage': grpc.unary_unary_rpc_method_handler(
                    servicer.GetMessage,
                    request_deserializer=database__pb2.SelectRequest.FromString,
                    response_serializer=database__pb2.SelectResponse.SerializeToString,
            ),
            'GetUnviewedMessagesForReminders': grpc.unary_unary_rpc_method_handler(
                    servicer.GetUnviewedMessagesForReminders,
                    request_deserializer=database__pb2.GetUnviewedMessagesRequest.FromString,
                    response_serializer=database__pb2.GetUnviewedMessagesResponse.SerializeToString,
            ),
            'LogReminderSent': grpc.unary_unary_rpc_method_handler(
                    servicer.LogReminderSent,
                    request_deserializer=database__pb2.LogReminderRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'GetReminderHistory': grpc.unary_unary_rpc_method_handler(
                    servicer.GetReminderHistory,
                    request_deserializer=database__pb2.GetReminderHistoryRequest.FromString,
                    response_serializer=database__pb2.GetReminderHistoryResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'databasepb.dbService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('databasepb.dbService', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class dbService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Select(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/Select',
            database__pb2.SelectRequest.SerializeToString,
            database__pb2.SelectResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def Insert(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/Insert',
            database__pb2.InsertRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetMessage(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/GetMessage',
            database__pb2.SelectRequest.SerializeToString,
            database__pb2.SelectResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetUnviewedMessagesForReminders(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/GetUnviewedMessagesForReminders',
            database__pb2.GetUnviewedMessagesRequest.SerializeToString,
            database__pb2.GetUnviewedMessagesResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def LogReminderSent(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/LogReminderSent',
            database__pb2.LogReminderRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetReminderHistory(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/databasepb.dbService/GetReminderHistory',
            database__pb2.GetReminderHistoryRequest.SerializeToString,
            database__pb2.GetReminderHistoryResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
