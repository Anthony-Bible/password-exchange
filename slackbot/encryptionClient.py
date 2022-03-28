import uuid
import os
import base64
from google.protobuf import field_mask_pb2
import grpc
# gazelle:ignore database
from protos import encryption_pb2
from protos import encryption_pb2_grpc
import database
SERVER_ADDRESS = os.environ['PASSWORDEXCHANGE_ENCRYPTIONSERVICE']
PORT = 50051

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
            print(encryptionbytes)
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

        db = database.databaseServiceClient()

        try:    
            print("encryting text")
            print(type(plaintext))            
            request.Key = self.generate_random_strng(32).encryptionbytes
            print(request)
            guid = uuid.uuid4().hex
            encrypt_response = self.stub.encryptMessage(request)
            
            insert_request = {'Uuid': guid, 'Content': encrypt_response.Ciphertext}
            print(insert_request)
            db.insert_message(insert_request)
            print("inserted into  database")
            return request.Key, str(guid)

        except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member

