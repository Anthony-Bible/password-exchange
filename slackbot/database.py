
import grpc

import database_pb2
import database_pb2_grpc

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
        self.channel = grpc.insecure_channel(f'{SERVER_ADDRESS}:{PORT}')
        self.stub = database_pb2_grpc.databaseServiceStub(self.channel)
    def generate_random_strng(self, length):
        """Generates a cryptographically rand string from the given length

        Args:
            length (int): Length of string you want generated.
        """
        try:
            random_request = encryption_pb2.Randomrequest(RandomLength=length)
            encryptionbytes = self.stub.GenerateRandomString(random_request)
            return encryptionbytes
        except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member

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
             response = self.stub.Insert(request)  
          except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member
        else:
            raise Exception("Request isn't right, it should have both a UUID and Content fields")