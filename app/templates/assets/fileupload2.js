// Takes a file, encrypts it, and uploads it to the server all in chunks
async function uploadFile(file, cipher) {
  cipher.start({ iv });
  // split file into chunks and spin up a promise for each chunk
    // Chunk size (in bytes)
  const chunkSize = 1024 * 1024 * 5;

  // The file ID returned after uploading each chunk
  let fileId = null;

  // Total number of chunks
  const numChunks = Math.ceil(file.size / chunkSize);
  console.log(`Total number of chunks: ${numChunks}`);
  let start = 0;
  let end = Math.min(chunkSize, file.size);
  let json;
  if (numchunks > 0){
      start = 0
          const formData = new FormData(document.querySelector('#contact-form'));
    json = await uploadChunk(file.slice(start, end), 1, numChunks,cipher, fileId, formData);

    console.log(end);
    console.log(file.slice(start, end).size);
  }
  // if there is only one chunk just return the json.URL
  if (numChunks === 1) {
    return json.URL;
  }
    let totalSize = 0;
  totalSize += file.slice(start, end).size;
  fileId = json.fileId;
   // if file id is empty something went wrong, throw an error and return nothing
  if (fileid === null) {
      console.log("fileid is null");
      return '';
    }
  const promises = [];
  for (let i = 2; i < numberOfChunks; i++) {
    const start = end;
    const end = Math.min( start + chunkSize, file.size);
    const chunk = file.slice(start, end);
    totalSize += chunk.size;

    promises.push(uploadChunk(chunk, i, numChunks, cipher, fileId ));
  }
    await Promise.all(promises);
    start = end;
    end = Math.min(start + chunkSize, file.size);
    totalSize += file.slice(start, end).size;
    json = await uploadChunk(file.slice(start), numChunks, numChunks, cipher, fileId );
    return json.URL;

}
async function uploadChunk(chunk, chunkNumber, totalChunks, cipher, fileId = null, formData = null) {

    // encrypt the chunk
    cipher.update(forge.util.createBuffer(chunk));
    const encryptedChunk = cipher.finish();
    // create a form data object to send the encrypted chunk
    const formData = new FormData();
    formData.append('file', encryptedChunk);
    // send the encrypted chunk to the server
    const apiUrl = "https://dev.password.exchange"


}

// Form submit event handler
document.querySelector('#contact-form').addEventListener('submit', async (event) => {
  event.preventDefault();

  // Get the file input element
  const fileInput = document.querySelector('#file-to-upload');
  const submitButton = document.querySelector('#submitButton');
  const loadingBar = document.querySelector('#loading-bar');
  const successArea = document.querySelector('#wasitasuccess');
  loadingBar.style.display = 'block';
  submitButton.style.display = 'none';
  const key = forge.random.getBytesSync(32);
  const iv = forge.random.getBytesSync(16);
  const cipher = forge.cipher.createCipher('AES-CTR', key);
  const decryptURL = await uploadFile(fileInput.files[0], cipher);

  // Upload the file in chunks and update the progress bar
  // const decryptUrl = await uploadFile(fileInput.files[0]);

  loadingBar.style.display = 'none';
  submitButton.style.display = 'block';
  // disable submit button
  submitButton.disabled = true;
  // turn successArea text green
  successArea.style.color = 'green';
  // display fileid on successarea
  successArea.innerHTML = `Your file was uploaded successfully. <br> Your Decrypt URL is <a href="${decryptUrl}" target="_blank">${decryptUrl}</a>`;
  // Do something with the final file ID
  console.log('Response:', decryptUrl);
});
