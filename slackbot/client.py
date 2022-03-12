import uuid

from google.protobuf import field_mask_pb2
import grpc

import encryption_pb2
import encryption_pb2_grpc
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
        self.stub = example_pb2_grpc.EncryptionPhotoServiceStub(self.channel)
    def generate_random_strng(self, length):
        """Generates a cryptographically random string from the given length

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

    def encrypt_text(self, plaintext):
        """encrypts the input

        Arguments:
            plaintext (string): Text to encrypt
        
        Returns:
           Outputed encrypted text
        """
        request = encryption_pb2.User(
            plaintext=plaintext
        )

        try:
            
            request['key'] = self.generate_random_strng(32).Encryptionbytes
            guid = uuid.uuid4().hex
            encrypt_response = self.stub.EncryptMessage(request)
            insert_request = {'Uuid': guid, 'Content': encrypt_response.Content}
            database.insert(insert_request)
            return request.key, guid

        except grpc.RpcError as err:
            print(err.details()) #pylint: disable=no-member
            print('{}, {}'.format(err.code().name, err.code().value)) #pylint: disable=no-member