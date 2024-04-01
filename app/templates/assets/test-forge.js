// get files from upload button and print them to console

// encrypt a chunk of the file
async function encryptChunk(chunk, chunkNumber, numberOfChunks, cipher) {
  console.log(`Encrypting chunk ${chunkNumber} of ${numberOfChunks}`);
  cipher.update(forge.util.createBuffer(await chunk.arrayBuffer()));
  cipher.finish();
  const encrypted = cipher.output;
  const tag = cipher.mode.tag;

  const encryptedConverted = encrypted.tobase64();
  console.log(`Encrypted chunk ${chunkNumber} of ${numberOfChunks}: ${encryptedConverted}`);
  return encryptedConverted;
}
// upload chunk to file, return file id
// write file to temporary storage with iv, encryptedChunk content and tag
async function uploadChunk(encryptedChunk, chunkNumber, numberOfChunks) {
  console.log(`Uploading chunk ${chunkNumber} of ${numberOfChunks}`);
}
// encrypt file in chunks
async function encryptFile(file) {
  const key = forge.random.getBytesSync(32);
  const iv = forge.random.getBytesSync(16);
  const cipher = forge.cipher.createCipher('AES-CTR', key);
  // use promises to split the file and encrypt each chunk asynchronously
  // this is to demonstrate that it is possible to encrypt large files in chunks
  cipher.start({ iv });
  // split file into chunks and spin up a promise for each chunk
  // Chunk size (in bytes)
  const chunkSize = 1024 * 1024 * 5;

  // The file ID returned after uploading each chunk
  const fileId = null;

  // Total number of chunks
  const numChunks = Math.ceil(file.size / chunkSize);

  let start = 0;
  let end = Math.min(chunkSize, file.size);
  console.log(`Total number of chunks: ${numChunks}`);
  let json;
  if (numChunks > 0) {
    start = 0;
  }
  // if there is only one chunk just return the json.URL
  if (numChunks === 1) {
    // print out the encrypted file base64 string
    const encrypted = await encryptChunk(file.slice(start), 1, 1, cipher);
    console.log('Encrypted file base64 string: ', encrypted);
    return '';
  }
  let totalSize = 0;
  totalSize += file.slice(start, end).size;
  // if file id is empty something went wrong, throw an error and return nothing
  const promises = [];
  for (let i = 2; i < numChunks; i += 1) {
    start = end;
    end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);
    totalSize += chunk.size;

    promises.push(encryptChunk(chunk, i, numChunks, cipher));
  }
  await Promise.all(promises);
  start = end;
  end = Math.min(start + chunkSize, file.size);
  totalSize += file.slice(start, end).size;
  json = await encryptChunk(file.slice(start), numChunks, numChunks, cipher);
  return json.URL;
}

document.querySelector('#contact-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  // get files from the input element
  const { files } = document.querySelector('[type=file]');
  // get the first file
  const file = files[0];
  // print the name of the fille
  console.log(file.name);
  // upload file in chunks
  encryptFile(file);
});
