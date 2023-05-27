// Require the Forge library.
const CHUNK_SIZE = 64 * 1024; // 64 KB

// Create an encryption key.
const key = forge.random.getBytesSync(16);

// Function to encrypt a chunk of data.
function encryptChunk(chunk, chunkNumber, totalChunks) {
  // Create a new, random IV for this chunk.
  const iv = forge.random.getBytesSync(12); // 96 bits is a common size for GCM.jjjj
  // print base64 key
  const cipher = forge.cipher.createCipher('AES-GCM', key);
  // trhows error if cipher.err
  if (cipher.err) {
    throw new Error(`Error creating cipher: ${cipher.err}`);
  }
  // Print key, iv, and chunkNumber before starting the cipher
  console.log(`Encrypting chunk number ${chunkNumber} out of ${totalChunks}`);
  console.log('Key:', forge.util.encode64(key));
  console.log('IV:', forge.util.encode64(iv));

  cipher.start({ iv, additionalData: `binary-encoded string${chunkNumber.toString()}${totalChunks.toString()}`, tagLength: 128 }); // optional additional data and tag length
  const buffer = forge.util.createBuffer(chunk, 'binary');
  cipher.update(buffer);
  cipher.finish();

  const encrypted = cipher.output;
  const { tag } = cipher.mode; // authentication tag
  // print size of tag
  console.log('Encrypted chunk size:', encrypted.length());
  console.log('Tag:', forge.util.bytesToHex(tag));
  // combine IV, encrypted data and tag
  const combined = forge.util.createBuffer();
  combined.putBytes(iv);
  combined.putBuffer(encrypted);
  combined.putBuffer(tag);
  // get size of chunk and print it

  return combined.getBytes();
}

// Read file in chunks.
function readFileInChunks(file) {
  let offset = 0;
  const reader = new FileReader();
  const outputBuffer = forge.util.createBuffer();
  let currentChunk = 1;
  // print ffile size
  console.log('file size: ', file.size);
  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
  reader.onload = function () {
    if (reader.error) {
      console.error('An error occurred while reading the file:', reader.error);
      return;
    }
    // Encrypt the current chunk.
    const encryptedChunk = encryptChunk(reader.result, currentChunk, totalChunks);

    // Append encrypted chunk to the output buffer.
    outputBuffer.putBuffer(forge.util.createBuffer(encryptedChunk));

    // Continue reading the next chunk if any.
    offset += CHUNK_SIZE;
    if (offset < file.size) {
      currentChunk += 1;
      readNextChunk();
    } else {
      // File has been read completely, now the 'outputBuffer' contains the encrypted file.
      // Convert to Blob and download it, for example.
      // log  current chunk and total chunks

      console.log(currentChunk, totalChunks);
      // log size of outputbuffer
      console.log('size of outputbuffer: ', outputBuffer.getBytes);
      const binaryString = outputBuffer.getBytes();
      const uint8Array = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        uint8Array[i] = binaryString.charCodeAt(i) & 0xff;
      }

      // Create Blob from Uint8Array
      const encryptedFile = new Blob([uint8Array], { type: 'application/octet-stream' });

      const url = URL.createObjectURL(encryptedFile);
      const link = document.createElement('a');
      link.href = url;
      link.download = 'encryptedFile.bin';
      link.click();
      // append key base64 encoded to url as hash
      window.location.hash = forge.util.encode64(key);
    }
  };

  function readNextChunk() {
    const slice = file.slice(offset, offset + CHUNK_SIZE);
    reader.readAsArrayBuffer(slice);
  }

  readNextChunk();
}
// Decrypt file in chunks.
//
// Function to decrypt a chunk of data.
function decryptChunk(chunkWithIvAndTag, decryptKeyDecode, chunkNumber, totalChunks) {
  // Extract IV, encrypted data, and tag.
  const iv = chunkWithIvAndTag.slice(0, 12);
  const tag = chunkWithIvAndTag.slice(-16);
  const chunk = chunkWithIvAndTag.slice(12, -16);
  // print size of chunk
  console.log('size of chunk: ', chunkWithIvAndTag.byteLength);
  // print size of iv and tag
  console.log('size of iv: ', iv.byteLength);
  console.log('size of tag: ', tag.byteLength);
  console.log(`Decrypting chunk number ${chunkNumber} out of ${totalChunks}`);
  console.log('Key:', forge.util.encode64(decryptKeyDecode));
  console.log('IV:', forge.util.encode64(iv));
  console.log('Tag:', forge.util.encode64(tag));
  // print base64 encoded iv and tag
  // console.log('IV: ', tag.toHex());
  console.log('tag:', forge.util.encode64(new Uint8Array(tag).join(',')));

  const decipher = forge.cipher.createDecipher('AES-GCM', decryptKeyDecode);
  decipher.start({
    iv, additionalData: `binary-encoded string${chunkNumber.toString()}${totalChunks.toString()}`, tagLength: 128, tag,
  });

  const buffer = forge.util.createBuffer(chunk, 'binary');
  decipher.update(buffer);

  if (decipher.finish()) {
    console.log('Decryption success!');
  } else {
    throw new Error('Authentication failed. The message is not authentic!');
  }

  return {
    data: decipher.output.getBytes(),
    chunkNumber,
  };
}

// Read file in chunks.
function decryptFileinChunks(file, decryptKeyDecode) {
  // This should be chunk size + iv size + tag size
  const ENCRYPTED_CHUNK_SIZE = CHUNK_SIZE + 12 + 16;
  console.log('encrypted chunk size: ', ENCRYPTED_CHUNK_SIZE);
  let offset = 0;
  const reader = new FileReader();
  const outputBuffer = forge.util.createBuffer();
  const totalChunks = Math.ceil(file.size / ENCRYPTED_CHUNK_SIZE);
  console.log(totalChunks);
  let currentChunk = 1;

  reader.onload = function () {
    if (reader.error) {
      console.error('An error occurred while reading the file:', reader.error);
      return;
    }

    // Convert ArrayBuffer to binary string
    //    const chunkAsBinaryString = forge.util.binary.raw.encode(new Uint8Array(reader.result));

    // Decrypt the current chunk.
    const decryptionResult = decryptChunk(reader.result, decryptKeyDecode, currentChunk, totalChunks);
    currentChunk = decryptionResult.chunkNumber + 1;

    // Append decrypted chunk to the output buffer.
    outputBuffer.putBytes(decryptionResult.data);

    // Continue reading the next chunk if any.
    offset += ENCRYPTED_CHUNK_SIZE;
    if (offset < file.size) {
      readNextChunk();
    } else {
      // File has been read completely, now the 'outputBuffer' contains the encrypted file.
      const binaryString = outputBuffer.getBytes();
      const uint8Array = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i += 1) { // eslint-disable-line no-plusplus
        uint8Array[i] = binaryString.charCodeAt(i) & 0xff;
      }
      // Convert to Blob and download it, for example.
      const decryptedFile = new Blob([uint8Array], { type: 'application/octet-stream' });
      const url = URL.createObjectURL(decryptedFile);
      const link = document.createElement('a');
      link.href = url;
      link.download = 'decryptedFile.bin';
      link.click();
    }
  };
  function readNextChunk() {
    const slice = file.slice(offset, offset + ENCRYPTED_CHUNK_SIZE);
    reader.readAsArrayBuffer(slice);
  }
  readNextChunk();
}
// Suppose 'file' is a File object you got from an <input> element or a Drag-and-Drop event.

document.querySelector('#contact-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  // get files from the input element
  const { files } = document.querySelector('[type=file]');
  // get the first file
  const file = files[0];
  // print the name of the fille
  console.log(file.name);
  // upload file in chunks
  readFileInChunks(file);
});

// prevent default when clicking the upload button
document.querySelector('#decrypt-button').addEventListener('click', (event) => {
  event.preventDefault();
  // get key from anchor text url
  const decryptKey = window.location.hash.substring(1);
  console.log(decryptKey);
  const decryptKeyDecode = forge.util.decode64(decryptKey);
  const { files } = document.querySelector('[type=file]');
  // get the first file
  const file = files[0];
  // print the name of the fille
  console.log(file.name);
  // upload file in chunks
  decryptFileinChunks(file, decryptKeyDecode);
});
