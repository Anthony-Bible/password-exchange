import uuid
import os
import base64
from google.protobuf import field_mask_pb2
from google.protobuf import empty_pb2
import grpc
#gazelle:resolve py encryption_pb2 //protos:encryption_pb2
#gazelle:resolve py database_pb2_grpc //protos:encryption_pb2_grpc
#gazelle:resolve py protos //protos:encryption_pb2_grpc

from protos import encryption_pb2
from protos import encryption_pb2_grpc
#gazelle:ignore database
import database
SERVER_ADDRESS = os.environ['PASSWORDEXCHANGE_ENCRYPTIONSERVICE']
PORT = 50051
db = database.databaseServiceClient()

class EncryptionServiceClient(object):
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
        self.stub = encryption_pb2_grpc.messageServiceStub(self.channel)
    def generate_random_strng(self, length):
        """Generates a cryptographically random string from the given length

        Args:
            length (int): Length of string you want generated.
        """
        try:
            random_request = encryption_pb2.Randomrequest(randomLength=length)
            encryptionbytes = self.stub.GenerateRandomString(random_request)
            return encryptionbytes
        except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member

    def encrypt_text(self, plaintext):
        """encrypts the input

        Arguments:
            plaintext (string): Text to encrypt
        
        Returns:
           Outputed encrypted text
        """
        request = encryption_pb2.EncryptedMessageRequest()
        request.PlainText.append(plaintext)


        try:    
            request.Key = self.generate_random_strng(32).encryptionbytes
            guid = uuid.uuid4().hex
            encrypt_response = self.stub.encryptMessage(request)
            for i in encrypt_response.Ciphertext:
                insert_request = {'uuid': guid, 'content': i}
                db.insert_message(insert_request)
            return request.Key, str(guid)

        except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member

