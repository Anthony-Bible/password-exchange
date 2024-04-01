const form = document.getElementById('contact-form');
const url = 'https://dev.password.exchange/';
let chunkCounter = 1;
const CHUNK_SIZE = 1024 * 1024 * 5; // 5MB chunk sizes
let fileId = '';
let start = 0;
let file;
let filename;
let numberofChunks; // Added these variable declarations

// Create an encryption key.
const key = forge.random.getBytesSync(16);

function encryptChunk(chunk, chunkNumber, totalChunks) {
  const iv = forge.random.getBytesSync(12);
  const cipher = forge.cipher.createCipher('AES-GCM', key);
  try {
    console.log(`Encrypting chunk number ${chunkNumber} out of ${totalChunks}`);
    console.log('Key:', forge.util.encode64(key));
    console.log('IV:', forge.util.encode64(iv));

    cipher.start({ iv, additionalData: `binary-encoded string${chunkNumber.toString()}${totalChunks.toString()}`, tagLength: 128 });
    const buffer = forge.util.createBuffer(chunk, 'binary');
    cipher.update(buffer);
    cipher.finish();
  } catch (error) {
    throw new Error(`Error creating cipher: ${error.message}`);
  }

  const encrypted = cipher.output;
  const { tag } = cipher.mode;
  console.log('Encrypted chunk size:', encrypted.length());
  console.log('Tag:', forge.util.bytesToHex(tag));
  const combined = forge.util.createBuffer();
  combined.putBytes(iv);
  combined.putBuffer(encrypted);
  combined.putBuffer(tag);

  return combined.getBytes();
}

function uploadChunk(chunkForm, chunkStart, chunkEnd) {
  const oReq = new XMLHttpRequest();
  oReq.upload.addEventListener('progress', updateProgress);
  oReq.open('POST', url, true);
  const blobEnd = chunkEnd - 1;
  const contentRange = `bytes ${chunkStart}-${blobEnd}/${file.size}`;
  oReq.setRequestHeader('Content-Range', contentRange);
  console.log('Content-Range', contentRange);

  function updateProgress(oEvent) {
    if (oEvent.lengthComputable) {
      const percentComplete = Math.round((oEvent.loaded / oEvent.total) * 100);
      const totalPercentComplete = Math.round(((chunkCounter - 1) / numberofChunks) * 100 + percentComplete / numberofChunks);
      document.getElementById('chunk-information').innerHTML = `Total uploaded: ${totalPercentComplete}%`;
    } else {
      console.log('not computable');
    }
  }

  oReq.onload = function uploadOnLoad(oEvent) {
    console.log('uploaded chunk');
    console.log('oReq.response', oReq.response);
    const resp = JSON.parse(oReq.response);
    fileId = resp.fileID;
    console.log('fileId', fileId);

    start += CHUNK_SIZE;
    if (start < file.size) {
      createChunk(fileId, start, file, numberofChunks); // Passing all arguments
    } else {
      console.log('all uploaded! Watch here: ', resp);
      document.getElementById('contact-form').innerHTML = `<a href="${resp.URL}">Click here to view your content</a>`;
    }
  };

  oReq.send(chunkForm);
}

function createChunk(sentChunkfileId, chunkStart, file, numberofChunks) {
  const outputBuffer = forge.util.createBuffer();
  console.log('created chunk: ', chunkCounter);
  const chunkEnd = Math.min(chunkStart + CHUNK_SIZE, file.size);
  const chunk = file.slice(chunkStart, chunkEnd);
  const encryptedChunk = encryptChunk(chunk, chunkCounter, numberofChunks);
  console.log(`i created a chunk of video ${chunkStart} - ${chunkEnd} minus 1`);
  const chunkForm = new FormData();
  chunkForm.append('totalChunks', numberofChunks);
  chunkForm.append('currentChunk', chunkCounter);

  if (sentChunkfileId.length > 0) {
    chunkForm.append('fileID', sentChunkfileId);
    console.log('added fileId');
  } else {
    // only post form elements on first post,
    // otherwise we'll associate it with previous upload (see above)
    chunkForm.append(
      'firstname',
      document.getElementById('firstname').value,
    );
    chunkForm.append('email', document.getElementById('email').value);
    chunkForm.append(
      'other_firstname',
      document.getElementById('other_firstname').value,
    );
    chunkForm.append(
      'other_lastname',
      document.getElementById('other_lastname').value,
    );
    chunkForm.append(
      'other_email',
      document.getElementById('other_email').value,
    );
    chunkForm.append('color', document.getElementById('color').value);
    chunkForm.append(
      'skipEmail',
      document.getElementById('skipEmail').value,
    );
    chunkForm.append(
      'form_message',
      document.getElementById('form_message').value,
    );
  }
  const blob = new Blob([encryptedChunk], { type: 'application/octet-stream' });
  chunkForm.append('file', blob, filename);
  console.log('added file');

  uploadChunk(chunkForm, chunkStart, chunkEnd);
  chunkCounter += 1;
}

form.addEventListener('submit', (event) => {
  event.preventDefault();
  const fileinput = document.getElementById('file-to-upload');
  if (fileinput.value) {
    file = fileinput.files[0]; // Setting the global file variable
    filename = fileinput.files[0].name; // Setting the global filename variable
    numberofChunks = Math.ceil(file.size / CHUNK_SIZE); // Setting the global numberofChunks variable
    console.log(`There will be ${numberofChunks}chunks uploaded`);
    const chunkEnd = start + CHUNK_SIZE;
    createChunk(fileId, start, file, numberofChunks); // Passing all arguments
  } else {
    console.log('no files selected');
  }
});
