const form = document.getElementById('contact-form');
// changed to sandbox, becuase we cannot have nice things
const url = 'https://dev.password.exchange/';
let chunkCounter = 0;
const CHUNK_SIZE = 1024 * 1024 * 5; // 1MB chunk sizes
let fileId = '';
const ChunkfileId = '';
// We may need this for decrypt url
// const playerUrl = '';
// TODO: We may needa to listen for another event

function uploadChunk(chunkForm, startUploadChunk, UploadchunkEnd) {
  const oReq = new XMLHttpRequest();
  oReq.upload.addEventListener('progress', updateProgress);
  oReq.open('POST', url, true);
  const blobEnd = UploadchunkEnd - 1;
  const contentRange = `bytes ${startUploadChunk}-${blobEnd}/${file.size}`;
  oReq.setRequestHeader('Content-Range', contentRange);
  console.log('Content-Range', contentRange);
  function updateProgress(oEvent) {
    if (oEvent.lengthComputable) {
      const percentComplete = Math.round(
        (oEvent.loaded / oEvent.total) * 100,
      );

      const totalPercentComplete = Math.round(
        ((chunkCounter - 1) / numberofChunks) * 100
                        + percentComplete / numberofChunks,
      );
      document.getElementById('chunk-information').innerHTML = `Total uploaded: ${totalPercentComplete}%`;
      //      console.log (percentComplete);
      // ...
    } else {
      console.log('not computable');
      // Unable to compute progress information since the total size is unknown
    }
  }
  oReq.onload = function uploadOnLoad(oEvent) {
    // Uploaded.
    console.log('uploaded chunk');
    console.log('oReq.response', oReq.response);
    const resp = JSON.parse(oReq.response);
    fileId = resp.fileID;
    // playerUrl = resp.assets.player;
    console.log('fileId', fileId);

    // now we have the video ID - loop through and add the remaining chunks
    // we start one chunk in, as we have uploaded the first one.
    // next chunk starts at + chunkSize from start
    start += CHUNK_SIZE;
    // if start is smaller than file size - we have more to still upload
    if (start < file.size) {
      // create the new chunk
      createChunk(fileId, startUploadChunk);
    } else {
      // the video is fully uploaded. there will now be a url in the response
      // playerUrl = resp.assets.player;
      console.log('all uploaded! Watch here: ', resp);
      document.getElementById('contact-form').innerHTML = `<a href="${resp.URL}">Click here to view your content</a>`;
    }
  };
  oReq.send(chunkForm);
}
function createChunk(sentChunkfileId, chunkStart, file, numberofChunks) {
  if (sentChunkfileId == null) {
    sentChunkfileId = '';
  }
  chunkCounter += 1;
  console.log('created chunk: ', chunkCounter);
  const chunkEnd = Math.min(chunkStart + CHUNK_SIZE, file.size);
  const chunk = file.slice(chunkStart, chunkEnd);
  console.log(
    `i created a chunk of video ${
      chunkStart
    } - ${
      chunkEnd
    } minus 1    `,
  );
  const chunkForm = new FormData();
  chunkForm.append('totalChunks', numberofChunks);
  chunkForm.append('currentChunk', chunkCounter);
  if (sentChunkfileId.length > 0) {
    // we have a fileId
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

  chunkForm.append('file', chunk, filename);
  console.log('added file');

  // created the chunk, now upload iit
  uploadChunk(chunkForm, chunkStart, chunkEnd);
}

form.addEventListener('submit', (event) => {
  event.preventDefault();
  const fileinput = document.getElementById('file-to-upload');
  if (fileinput.value) {
    console.log(fileinput.length);
    const file = fileinput.files[0];
    const filename = fileinput.files[0].name;
    const numberofChunks = Math.ceil(file.size / CHUNK_SIZE);
    console.log(`There will be ${numberofChunks}chunks uploaded`);
    const start = 0;
    const chunkEnd = start + CHUNK_SIZE;
    // upload the first chunk to get the fileId
    createChunk(fileId, start);
  } else {
    console.log('no files selected');
  }
  // get the file name to name the file.  If we do not name the file,
  // the upload will be called 'blob'
});
