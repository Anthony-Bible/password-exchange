/* jshint esversion: 8, browser: true, devel: true, globalstrict: true */

// file-upload.js — encrypted chunked file upload/download via the
// POST /api/v1/files/initiate → POST /api/v1/files/:fileID/chunks →
// GET /api/v1/files/:fileID (X-File-Key: <encodedKey> header) backend.
//
// Encryption is handled server-side (AES-256-GCM). The client sends
// plaintext chunks and receives a base64url-encoded key from the initiate
// response. That key is embedded in the share URL fragment (#k=...) so the
// downloader's browser reads it client-side and sends it as the X-File-Key header.

'use strict';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// Default chunk size: 5 MiB (must fit within the server's multipart memory
// limit and leave room for the GCM overhead on the stored side).
const FILE_UPLOAD_CHUNK_SIZE = 5 * 1024 * 1024;
const FILE_UPLOAD_CHUNK_CONCURRENCY = 4;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// formatFileSize returns a human-readable size string.
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    return (bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1) + ' ' + units[i];
}

// getMimeType returns the file's MIME type, falling back to a safe default.
function getMimeType(file) {
    return file.type || 'application/octet-stream';
}

// showFileUploadError displays an error inside the file upload section.
function showFileUploadError(container, message) {
    const errorEl = container.querySelector('.file-upload-error');
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.style.display = 'block';
    }
}

// hideFileUploadError clears any displayed error.
function hideFileUploadError(container) {
    const errorEl = container.querySelector('.file-upload-error');
    if (errorEl) {
        errorEl.textContent = '';
        errorEl.style.display = 'none';
    }
}

// updateProgress sets the Bootstrap progress bar to a given percentage.
function updateProgress(container, percent, label) {
    const bar = container.querySelector('.file-upload-progress-bar');
    const status = container.querySelector('.file-upload-status');
    if (bar) {
        bar.style.width = percent + '%';
        bar.setAttribute('aria-valuenow', percent);
        bar.textContent = Math.round(percent) + '%';
    }
    if (status && label) {
        status.textContent = label;
    }
}

// resetProgress hides the progress section.
function resetProgress(container) {
    const section = container.querySelector('.file-upload-progress-section');
    if (section) {
        section.style.display = 'none';
    }
    updateProgress(container, 0, '');
}

// showProgress makes the progress section visible.
function showProgress(container) {
    const section = container.querySelector('.file-upload-progress-section');
    if (section) {
        section.style.display = 'block';
    }
}

// ---------------------------------------------------------------------------
// Core upload logic
// ---------------------------------------------------------------------------

/**
 * Processes every item in `items` by calling `fn(item)`, keeping at most
 * `limit` concurrent promises in flight.
 *
 * Rejects as soon as any `fn` call rejects, and stops scheduling new items
 * after the first failure.
 */
async function runWithConcurrency(items, fn, limit, onError) {
    if (!limit || limit < 1) limit = 1;

    const iter = items[Symbol.iterator]();
    let stopped = false;

    async function worker() {
        for (;;) {
            if (stopped) return;
            const { value, done } = iter.next();
            if (done) return;
            try {
                await fn(value);
            } catch (err) {
                stopped = true;
                if (onError) {
                    try {
                        onError(err);
                    } catch (_) {
                        // Ignore onError callback failures; preserve the original error.
                    }
                }
                throw err;
            }
        }
    }

    const workerCount = Math.min(limit, Array.isArray(items) ? items.length : limit);
    await Promise.all(Array.from({ length: workerCount }, () => worker()));
}

/**
 * uploadFile drives the full chunked upload flow:
 *   1. POST /api/v1/files/initiate  → { fileID, sessionID, encodedKey }
 *   2. POST /api/v1/files/:fileID/chunks  (one per chunk, 1-based index)
 * Returns { fileID, encodedKey } on success, throws on failure.
 *
 * @param {File} file          The File object selected by the user.
 * @param {string} messageID   The parent message ID returned by the message API.
 * @param {Function} onProgress  Called with (chunksDone, totalChunks).
 * @param {AbortSignal} [signal] Optional AbortSignal for cancellation.
 */
async function uploadFile(file, messageID, onProgress, signal) {
    const chunkSize = FILE_UPLOAD_CHUNK_SIZE;
    const totalSize = file.size;
    const totalChunks = Math.max(1, Math.ceil(totalSize / chunkSize));

    // --- Step 1: Initiate upload ---
    const initiateResp = await fetch('/api/v1/files/initiate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            filename: file.name,
            contentType: getMimeType(file),
            totalSize: totalSize,
            chunkSize: chunkSize,
            messageID: messageID,
        }),
        signal,
    });

    if (!initiateResp.ok) {
        const body = await initiateResp.json().catch(() => ({}));
        throw new Error(body.message || 'Failed to initiate file upload (' + initiateResp.status + ')');
    }

    const { fileID, sessionID, encodedKey } = await initiateResp.json();

    if (!fileID || !sessionID || !encodedKey) {
        throw new Error('Unexpected response from file initiation: missing fileID, sessionID, or encodedKey');
    }

    // --- Step 2: Upload chunks (parallel, capped at FILE_UPLOAD_CHUNK_CONCURRENCY) ---
    let completedChunks = 0;
    let uploadFailed = false;

    // Internal controller lets us cancel all in-flight chunk fetches the moment
    // any single chunk fails, avoiding wasted bandwidth against a dead session.
    const chunkAbortController = new AbortController();
    const chunkSignal = signal
        ? (AbortSignal.any
            ? AbortSignal.any([signal, chunkAbortController.signal])
            : chunkAbortController.signal)
        : chunkAbortController.signal;

    const chunkIndices = Array.from({ length: totalChunks }, (_, i) => i + 1);

    await runWithConcurrency(chunkIndices, async (chunkIndex) => {
        const start = (chunkIndex - 1) * chunkSize;
        const end = Math.min(start + chunkSize, totalSize);
        const chunkBlob = file.slice(start, end);

        const formData = new FormData();
        formData.append('sessionID', sessionID);
        formData.append('chunkIndex', String(chunkIndex));
        formData.append('totalChunks', String(totalChunks));
        formData.append('data', chunkBlob, 'chunk.bin');

        const chunkResp = await fetch('/api/v1/files/' + encodeURIComponent(fileID) + '/chunks', {
            method: 'POST',
            body: formData,
            signal: chunkSignal,
        });

        if (!chunkResp.ok) {
            const body = await chunkResp.json().catch(() => ({}));
            if (chunkResp.status === 410) throw new Error('Upload session expired. Please try again.');
            if (chunkResp.status === 404) throw new Error('Upload session not found. Please try again.');
            throw new Error(body.message || 'Failed to upload chunk ' + chunkIndex + ' (' + chunkResp.status + ')');
        }

        completedChunks++;
        // Guard needed: in-flight workers can resolve *after* runWithConcurrency's
        // internal stopped flag is set, so we need uploadFailed to suppress stale
        // progress callbacks that would fire after a sibling chunk has already failed.
        if (!uploadFailed && onProgress) onProgress(completedChunks, totalChunks);
    }, FILE_UPLOAD_CHUNK_CONCURRENCY, () => {
        uploadFailed = true;
        chunkAbortController.abort();
    });

    return { fileID, encodedKey };
}

// ---------------------------------------------------------------------------
// Download helpers
// ---------------------------------------------------------------------------

/**
 * buildFileShareURL returns the shareable URL for a file download page.
 * The key is placed in the URL fragment (#k=...) so it is never sent to
 * the server in request logs or Referer headers. The recipient's browser
 * reads the fragment client-side and sends it as the X-File-Key header.
 *
 * Format: <origin>/files/<fileID>#k=<encodedKey>
 *
 * @param {string} fileID
 * @param {string} encodedKey  base64url-encoded 32-byte key (from InitiateUpload)
 * @param {string} [origin]    defaults to window.location.origin
 */
function buildFileShareURL(fileID, encodedKey, origin) {
    const base = (origin || window.location.origin) + '/files/';
    return base + encodeURIComponent(fileID) + '#k=' + encodedKey;
}

/**
 * buildCombinedShareURL appends file params to a message share URL so the
 * recipient can download the attached file from the same link.
 *
 * @param {string} messageWebUrl  The message share URL (e.g. result.webUrl)
 * @param {string} fileID
 * @param {string} encodedKey
 */
function buildCombinedShareURL(messageWebUrl, fileID, encodedKey) {
    const u = new URL(messageWebUrl, window.location.origin);
    u.hash = 'fid=' + encodeURIComponent(fileID) + '&fk=' + encodedKey;
    return u.toString();
}

/**
 * extractCombinedFileParams reads the fid/fk params embedded in the current
 * page's URL fragment by buildCombinedShareURL.
 *
 * Returns { fileID, encodedKey } or null if either value is absent.
 */
function extractCombinedFileParams() {
    const params = new URLSearchParams(window.location.hash.slice(1));
    const fileID = params.get('fid');
    const encodedKey = params.get('fk');
    if (!fileID || !encodedKey) return null;
    return { fileID: decodeURIComponent(fileID), encodedKey };
}

/**
 * extractFileDownloadParams parses the current page URL to recover the fileID
 * and encodedKey needed to call the download API.
 *
 * Expected URL: <origin>/files/<fileID>#k=<encodedKey>
 *
 * Returns { fileID, encodedKey } or null if either value is absent.
 */
function extractFileDownloadParams() {
    // fileID is the last non-empty path segment.
    const pathParts = window.location.pathname.replace(/\/$/, '').split('/');
    const fileID = decodeURIComponent(pathParts[pathParts.length - 1] || '');

    // Key is in the fragment as "k=<value>".
    const hash = window.location.hash.slice(1); // strip leading '#'
    const params = new URLSearchParams(hash);
    const encodedKey = params.get('k');

    if (!fileID || !encodedKey) {
        return null;
    }
    return { fileID, encodedKey };
}

/**
 * fetchAndSaveFile calls GET /api/v1/files/:fileID with the key in the
 * X-File-Key request header, streams the server-decrypted bytes back,
 * and triggers a browser save-as download.
 * The server performs all AES-256-GCM decryption; no client-side crypto needed.
 *
 * @param {string} fileID
 * @param {string} encodedKey  base64url-encoded key (with padding, using - and _)
 * @param {Function} [onProgress]  Called with (bytesReceived, totalBytes) during streaming
 */
async function fetchAndSaveFile(fileID, encodedKey, onProgress) {
    // Send the key as the X-File-Key request header so it never appears in
    // server access logs, proxy logs, browser history, or Referer headers.
    const apiURL = '/api/v1/files/' + encodeURIComponent(fileID);

    const response = await fetch(apiURL, {
        headers: { 'X-File-Key': encodedKey },
    });

    if (!response.ok) {
        let msg;
        if (response.status === 400) {
            msg = 'Invalid or missing decryption key. The link may be malformed.';
        } else if (response.status === 404) {
            msg = 'File not found. It may have been deleted or the link is incorrect.';
        } else if (response.status === 410) {
            msg = 'This file link has expired and is no longer available.';
        } else {
            const body = await response.json().catch(() => ({}));
            msg = body.message || 'Download failed (' + response.status + ').';
        }
        throw new Error(msg);
    }

    // Stream the response, reporting progress when Content-Length is available.
    const contentLength = parseInt(response.headers.get('Content-Length') || '0', 10);
    const reader = response.body.getReader();
    const parts = [];
    let received = 0;

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        parts.push(value);
        received += value.length;
        if (onProgress && contentLength > 0) {
            onProgress(received, contentLength);
        }
    }

    const blob = new Blob(parts, {
        type: response.headers.get('Content-Type') || 'application/octet-stream',
    });

    // Extract filename from Content-Disposition: attachment; filename="..."
    let filename = 'download';
    const disposition = response.headers.get('Content-Disposition') || '';
    const fnMatch = disposition.match(/filename="([^"]+)"/);
    if (fnMatch) {
        filename = fnMatch[1];
    }

    // Trigger the browser save-as dialog.
    const objectURL = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = objectURL;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(objectURL);
}

// ---------------------------------------------------------------------------
// UI wiring
// ---------------------------------------------------------------------------

/**
 * initializeFileUpload wires up the file upload UI inside `container`.
 * It attaches to the file picker, drives the upload on form submission, and
 * renders a copyable download link when the upload completes.
 *
 * @param {HTMLElement} container  The `.file-upload-section` wrapper element.
 * @param {Function} getMessageID  Async function that returns the message ID
 *                                 string for the current submission; called
 *                                 only when a file is selected.
 */
function initializeFileUpload(container, getMessageID) {
    const fileInput = container.querySelector('#file-upload-input');
    const fileLabel = container.querySelector('.file-upload-label-text');
    const fileSizeHint = container.querySelector('.file-upload-size-hint');
    const progressSection = container.querySelector('.file-upload-progress-section');
    const resultSection = container.querySelector('.file-upload-result-section');
    const resultUrl = container.querySelector('#file-upload-result-url');
    const copyBtn = container.querySelector('#file-upload-copy-btn');

    if (!fileInput) return;

    // Update the label when a file is chosen.
    fileInput.addEventListener('change', function () {
        hideFileUploadError(container);
        resetProgress(container);
        if (resultSection) resultSection.style.display = 'none';

        if (this.files && this.files[0]) {
            const f = this.files[0];
            if (fileLabel) fileLabel.textContent = f.name;
            if (fileSizeHint) fileSizeHint.textContent = formatFileSize(f.size);
        } else {
            if (fileLabel) fileLabel.textContent = 'Choose a file or drag & drop';
            if (fileSizeHint) fileSizeHint.textContent = '';
        }
    });

    // Copy-to-clipboard for the result URL.
    if (copyBtn && resultUrl) {
        copyBtn.addEventListener('click', async function () {
            try {
                await navigator.clipboard.writeText(resultUrl.value);
                const orig = this.innerHTML;
                this.innerHTML = '<i class="fas fa-check"></i>';
                this.classList.replace('btn-outline-secondary', 'btn-success');
                setTimeout(() => {
                    this.innerHTML = orig;
                    this.classList.replace('btn-success', 'btn-outline-secondary');
                }, 1500);
            } catch (_) {
                resultUrl.select();
                document.execCommand('copy');
            }
        });
    }

    // Expose an async function that the form-submit handler can await.
    // Returns { fileID, encodedKey, shareURL } on success, null if no file was
    // selected. Throws on error.
    // opts.showResult (default true) controls whether the built-in result URL
    // banner is shown; pass false when the caller embeds the URL elsewhere.
    container.runUpload = async function (messageID, opts) {
        const file = fileInput.files && fileInput.files[0];
        if (!file) return null;

        hideFileUploadError(container);
        showProgress(container);
        updateProgress(container, 0, 'Uploading…');

        try {
            const { fileID, encodedKey } = await uploadFile(
                file,
                messageID,
                function (done, total) {
                    const pct = (done / total) * 100;
                    updateProgress(container, pct, 'Uploading chunk ' + done + ' of ' + total + '…');
                }
            );

            updateProgress(container, 100, 'Upload complete');

            const shareURL = buildFileShareURL(fileID, encodedKey);

            if (resultUrl) resultUrl.value = shareURL;
            if (resultSection && (!opts || opts.showResult !== false)) resultSection.style.display = 'block';

            return { fileID, encodedKey, shareURL };
        } catch (err) {
            resetProgress(container);
            showFileUploadError(container, err.message || 'File upload failed.');
            throw err;
        }
    };
}

// ---------------------------------------------------------------------------
// Download page wiring
// ---------------------------------------------------------------------------

/**
 * initializeFileDownloadPage wires up the file download page.
 * It reads the fileID + key from the current URL (fragment-based share URL),
 * shows a download button, and drives the fetch+save flow when clicked.
 *
 * Expected URL: <origin>/files/<fileID>#k=<encodedKey>
 *
 * Call this from the DOMContentLoaded handler on the file-download template.
 *
 * @param {HTMLElement} container  The page wrapper element (id="file-download-container")
 */
function initializeFileDownloadPage(container) {
    const params = extractFileDownloadParams();

    const loadingEl    = container.querySelector('#download-loading');
    const readyEl      = container.querySelector('#download-ready');
    const progressEl   = container.querySelector('#download-progress-section');
    const progressBar  = container.querySelector('#download-progress-bar');
    const statusEl     = container.querySelector('#download-status');
    const errorEl      = container.querySelector('#download-error');
    const errorMsgEl   = container.querySelector('#download-error-message');
    const doneEl       = container.querySelector('#download-done');
    const downloadBtn  = container.querySelector('#download-start-btn');

    function showState(id) {
        [loadingEl, readyEl, progressEl, errorEl, doneEl].forEach(function (el) {
            if (el) el.style.display = 'none';
        });
        const target = container.querySelector('#' + id);
        if (target) target.style.display = 'block';
    }

    function setStatus(msg) {
        if (statusEl) statusEl.textContent = msg;
    }

    function showError(msg) {
        if (errorMsgEl) errorMsgEl.textContent = msg;
        showState('download-error');
    }

    // If params are missing, show an error immediately — the link is malformed.
    if (!params) {
        showError('This link is missing the file ID or decryption key. ' +
            'Please check the link and try again.');
        return;
    }

    // Show the ready state with the download button.
    showState('download-ready');

    if (downloadBtn) {
        downloadBtn.addEventListener('click', async function () {
            downloadBtn.disabled = true;
            downloadBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Downloading…';

            showState('download-progress-section');
            setStatus('Connecting…');

            try {
                await fetchAndSaveFile(
                    params.fileID,
                    params.encodedKey,
                    function (received, total) {
                        const pct = Math.round((received / total) * 100);
                        setStatus('Downloading… ' + pct + '%');
                        if (progressBar) {
                            progressBar.style.width = pct + '%';
                            progressBar.setAttribute('aria-valuenow', pct);
                            progressBar.textContent = pct + '%';
                        }
                    }
                );
                showState('download-done');
            } catch (err) {
                showError(err.message || 'Download failed. Please try again.');
            }
        });
    }
}
