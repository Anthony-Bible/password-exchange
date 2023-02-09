// Function to upload a chunk of a file
async function uploadChunk(chunk, chunkNumber, totalChunks, fileId = null, formData = null) {
  // API endpoint for uploading a chunk
  const apiUrl = 'https://dev.password.exchange';

  // Build the form data to be sent with the chunk
  if (!formData) {
    formData = new FormData();
  }

  formData.append('file', chunk);
  if (fileId) {
    formData.append('fileId', fileId);
  }
  formData.append('currentChunk', chunkNumber);
  formData.append('totalChunks', totalChunks);
  // Make the API call to upload the chunk
  const response = await fetch(apiUrl, {
    method: 'POST',
    body: formData,
  });

  // Return the file ID returned by the API
  const json = await response.json();
  return json;
}

// Function to upload a file in chunks
async function uploadFile(file) {
  // Chunk size (in bytes)
  const chunkSize = 1024 * 1024 * 5;

  // The file ID returned after uploading each chunk
  let fileId = null;

  // Total number of chunks
  const numChunks = Math.ceil(file.size / chunkSize);

  // Loop through all the chunks
  // Calculate the start and end bytes for the chunk

  // Upload the first chunk and get the file ID
  let json;
  if (numChunks > 0) {
    const start = 0;
    const end = Math.min(chunkSize, file.size);
    const chunk = file.slice(start, end);
    const formData = new FormData(document.querySelector('#contact-form'));
    json = await uploadChunk(chunk, 1, numChunks, fileId, formData);
  }
  if (numChunks === 1) {
    return json.URL;
  }
  fileId = json.fileId;

  // check if fileid is empty
  if (fileId === null) {
    console.log('fileId is null');
    return '';
  }
  // Upload the remaining chunks concurrently
  const promises = [];
  for (let i = 1; i < numChunks; i += 1) {
    const start = i * chunkSize;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);
    promises.push(uploadChunk(chunk, i, numChunks, fileId));
  }

  // Wait for all the promises to resolve
  await Promise.all(promises);

  // Return the final file ID
  return json.URL;
}

// Function to update the progress bar

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
  // Upload the file in chunks and update the progress bar
  const decryptUrl = await uploadFile(fileInput.files[0]);

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
