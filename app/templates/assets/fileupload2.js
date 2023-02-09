// Function to upload a chunk of a file
async function uploadChunk(chunk, fileId = null) {
  // API endpoint for uploading a chunk
  const apiUrl = 'https://dev.password.exchange';

  // Build the form data to be sent with the chunk
  const formData = new FormData();
  formData.append('file', chunk);
  if (fileId) {
    formData.append('fileId', fileId);
  }

  // Make the API call to upload the chunk
  const response = await fetch(apiUrl, {
    method: 'POST',
    body: formData,
  });

  // Return the file ID returned by the API
  const json = await response.json();
  fileId = json.fileId;

  // Return the file ID
  return fileId;
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
  if (numChunks > 0) {
    const start = 0;
    const end = Math.min(chunkSize, file.size);
    const chunk = file.slice(start, end);
    const json = await uploadChunk(chunk);
    fileId = json.fileId;
  }
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
    promises.push(uploadChunk(chunk, fileId));
  }

  // Wait for all the promises to resolve
  await Promise.all(promises);

  // Return the final file ID
  return fileId;
}

// Function to update the progress bar

// Form submit event handler
document.querySelector('#contact-form').addEventListener('submit', async (event) => {
  event.preventDefault();

  // Get the file input element
  const fileInput = document.querySelector('#file-to-upload');

  const loadingBar = document.querySelector('#loading-bar');
  loadingBar.style.display = 'block';

  // Upload the file in chunks and update the progress bar
  const fileId = await uploadFile(fileInput.files[0]);

  loadingBar.style.display = 'none';

  // Do something with the final file ID
  console.log('File ID:', fileId);
});
