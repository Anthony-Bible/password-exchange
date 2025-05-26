
import grpc
import os
import sys
import re
#gazelle:resolve py database_pb2 //protos:database_pb2
#gazelle:resolve py database_pb2_grpc //protos:database_pb2_grpc
from protos import database_pb2
from protos import database_pb2_grpc
from google.protobuf import empty_pb2
from google.protobuf import json_format
from google.protobuf.json_format import MessageToJson

import json


SERVER_ADDRESS = os.environ['PASSWORDEXCHANGE_DATABASESERVICE']
PORT = 8080
BLOCK_SIZE = 20000



class databaseServiceClient(object):
    def __init__(self):
        """Initializer. 
           Creates a gRPC channel for connecting to the server.
           Adds the channel to the generated client stub.
        Arguments:
            None.
        
        Returns:
            None.
        """
        self.channel = grpc.insecure_channel(f'{SERVER_ADDRESS}')
        self.stub = database_pb2_grpc.dbServiceStub(self.channel)
    def insert_message(self, request):
        """calls grpc function Insert and inserts UUID content

        Arguments:
           request (dict): {'uuid', 'content', 'passphrase', 'max_view_count'}
           All fields are required for InsertRequest
        
        Returns:
           None
        """
        required_fields = {'uuid', 'content', 'passphrase', 'max_view_count'}
        
        
        if required_fields <= request.keys():
          try:
             request_serialized = database_pb2.InsertRequest()
             json_format.Parse(json.dumps(request), request_serialized)
             self.stub.Insert(request_serialized)  
          except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member
        else:
            raise Exception("Request isn't right, it should have uuid, content, passphrase, and max_view_count fields")
