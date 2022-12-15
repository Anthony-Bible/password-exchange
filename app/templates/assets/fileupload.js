const input = document.querySelector('#file-to-upload')
//changed to sandbox, becuase we cannot have nice things
const url = 'https://ws.api.video/upload?token=to5PoOjCz98FdLGnsrFflnYo'
var chunkCounter = 0
//break into 100 MB chunks.
const chunkSize = 100000000
var fileId = ''
// We may need this for decrypt url
var playerUrl = ''
//TODO: We may needa to listen for another event
input.addEventListener('change', () => {
    const file = input.files[0]
    //get the file name to name the file.  If we do not name the file, the upload will be called 'blob'
    const filename = input.files[0].name
    var numberofChunks = Math.ceil(file.size / chunkSize)
    console.log('There will be ' + numberofChunks + 'chunks uploaded')
    var start = 0
    var chunkEnd = start + chunkSize
    //upload the first chunk to get the fileId
    createChunk(fileId, start)

    function createChunk(fileId, start, end) {
        chunkCounter++
        console.log('created chunk: ', chunkCounter)
        chunkEnd = Math.min(start + chunkSize, file.size)
        const chunk = file.slice(start, chunkEnd)
        console.log(
            'i created a chunk of video' + start + '-' + chunkEnd + 'minus 1    '
        )
        const chunkForm = new FormData()
        if (fileId.length > 0) {
            //we have a fileId
            chunkForm.append('fileId', fileId)
            console.log('added fileId')
        }
        chunkForm.append('file', chunk, filename)
        console.log('added file')

        //created the chunk, now upload iit
        uploadChunk(chunkForm, start, chunkEnd)
    }
    function uploadChunk(chunkForm, start, chunkEnd) {
        var oReq = new XMLHttpRequest()
        oReq.upload.addEventListener('progress', updateProgress)
        oReq.open('POST', url, true)
        var blobEnd = chunkEnd - 1
        var contentRange = 'bytes ' + start + '-' + blobEnd + '/' + file.size
        oReq.setRequestHeader('Content-Range', contentRange)
        console.log('Content-Range', contentRange)
        function updateProgress(oEvent) {
            if (oEvent.lengthComputable) {
                var percentComplete = Math.round(
                    (oEvent.loaded / oEvent.total) * 100
                )

                var totalPercentComplete = Math.round(
                    ((chunkCounter - 1) / numberofChunks) * 100 +
                        percentComplete / numberofChunks
                )
                document.getElementById('chunk-information').innerHTML =
                    'Total uploaded: ' +
                    totalPercentComplete +
                    '%'
                //      console.log (percentComplete);
                // ...
            } else {
                console.log('not computable')
                // Unable to compute progress information since the total size is unknown
            }
        }
        oReq.onload = function (oEvent) {
            // Uploaded.
            console.log('uploaded chunk')
            console.log('oReq.response', oReq.response)
            var resp = JSON.parse(oReq.response)
            fileId = resp.fileId
            //playerUrl = resp.assets.player;
            console.log('fileId', fileId)

            //now we have the video ID - loop through and add the remaining chunks
            //we start one chunk in, as we have uploaded the first one.
            //next chunk starts at + chunkSize from start
            start += chunkSize
            //if start is smaller than file size - we have more to still upload
            if (start < file.size) {
                //create the new chunk
                createChunk(fileId, start)
            } else {
                //the video is fully uploaded. there will now be a url in the response
                playerUrl = resp.assets.player
                console.log('all uploaded! Watch here: ', playerUrl)
            }
        }
        oReq.send(chunkForm)
    }
})
