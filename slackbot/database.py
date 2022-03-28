
import grpc
import os
import re
#gazelle:resolve py database_pb2 //protos:database_pb2
#gazelle:resolve py database_pb2_grpc //protos:database_pb2_grpc
from protos import database_pb2
from protos import database_pb2_grpc

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
           request (dict): {'Uuid', 'Content'}
        
        Returns:
           None
        """
        required_fields = {'Uuid', 'Content'}

        
        if required_fields <= request.keys() <= required_fields:
          try:
             print("inserting into database")
             response = self.stub.Insert(request)  
          except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member
        else:
            raise Exception("Request isn't right, it should have both a UUID and Content fields")
